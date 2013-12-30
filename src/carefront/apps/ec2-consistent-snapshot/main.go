package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"carefront/libs/aws"
	"carefront/libs/aws/ec2"
	"carefront/libs/cmd/cryptsetup"
	"carefront/libs/cmd/dmsetup"
	"carefront/libs/cmd/lvm"
	"carefront/libs/cmd/mount"
	"github.com/go-sql-driver/mysql"
)

var config = struct {
	// AWS
	AWSRole    string
	AWSKeys    aws.Keys
	Region     string
	InstanceId string
	// FS Freeze
	Filesystem string
	FreezeCmd  string
	Devices    []string // Used to lookup which EBS volumes to snapshot
	EBSVolumes []string // IDs for the EBS volumes
	// MySQL
	Config     string
	Host       string
	Port       int
	Socket     string
	Username   string
	Password   string
	CACert     string
	ClientCert string
	ClientKey  string

	db            *sql.DB
	awsAuth       aws.Auth
	ec2           *ec2.EC2
	ebsVolumeInfo []*ec2.Volume
}{
	FreezeCmd: "xfs_freeze",
	Host:      "127.0.0.1",
	Port:      3306,
	Socket:    "/var/run/mysqld/mysqld.sock",
	Username:  "root",
}

var cnfSearchPath = []string{
	"~/.my.cnf",
	"/etc/mysql/my.cnf",
}

func init() {
	flag.StringVar(&config.Filesystem, "fs", config.Filesystem, "Path to filesystem to freeze")
	flag.StringVar(&config.AWSRole, "role", config.AWSRole, "AWS Role")
	flag.StringVar(&config.Region, "region", config.Region, "EC2 Region")
	flag.StringVar(&config.Config, "mysql.config", config.Config, "Path to my.cnf")
	flag.StringVar(&config.Host, "mysql.host", config.Host, "MySQL host")
	flag.IntVar(&config.Port, "mysql.port", config.Port, "MySQL port")
	flag.StringVar(&config.Username, "mysql.user", config.Username, "MySQL username")
	flag.StringVar(&config.Password, "mysql.password", config.Password, "MySQL password")
}

func readMySQLConfig(path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()

	cnf, err := ParseConfig(fi)
	if err != nil {
		return err
	}

	for _, secName := range []string{"client", "ec2-consistent-snapshot"} {
		if sec := cnf[secName]; sec != nil {
			if port, err := strconv.Atoi(sec["port"]); err == nil {
				config.Port = port
			}
			if s := sec["host"]; s != "" {
				config.Host = s
			}
			if s := sec["socket"]; s != "" {
				config.Socket = s
			}
			if s := sec["ssl-ca"]; s != "" {
				config.CACert = s
			}
			if s := sec["ssl-cert"]; s != "" {
				config.ClientCert = s
			}
			if s := sec["ssl-key"]; s != "" {
				config.ClientKey = s
			}
			if s := sec["user"]; s != "" {
				config.Username = s
			}
			if s := sec["password"]; s != "" {
				config.Password = s
			}
		}
	}

	return nil
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if config.Filesystem == "" {
		log.Fatalf("Missing required option -fs")
	}

	// Resolve devices from the mount point. It may be LUKS and/or LVM.
	if len(config.Devices) == 0 {
		mounts, err := mount.Default.GetMounts()
		if err != nil {
			log.Fatalf("Failed to get mounts: %+v", err)
		}
		if mnt := mounts[config.Filesystem]; mnt == nil {
			log.Fatalf("Mount not found for path %s", config.Filesystem)
		} else {
			device := mnt.Device
			for {
				dev, err := dmsetup.Default.DMInfo(device)
				if err != nil {
					log.Fatalf("dmsetup info failed: %+v", err)
				}
				if strings.HasPrefix(dev.UUID, "CRYPT-LUKS") {
					cs, err := cryptsetup.Default.Status(device)
					if err != nil {
						log.Fatalf("cryptsetup status filed: %+v", err)
					}
					device = cs.Device
				} else if strings.HasPrefix(dev.UUID, "LVM-") {
					info, err := lvm.Default.LVS(device)
					if err != nil {
						log.Fatalf("lvs failed: %+v", err)
					}
					config.Devices = info.Devices
					break
				} else {
					config.Devices = []string{device}
					break
				}
			}
		}
	}

	if config.AWSRole != "" {
		if config.AWSRole == "*" {
			config.AWSRole = ""
		}
		cred, err := aws.CredentialsForRole(config.AWSRole)
		if err != nil {
			log.Fatal(err)
		}
		config.awsAuth = cred
	} else {
		if keys := aws.KeysFromEnvironment(); keys.AccessKey == "" || keys.SecretKey == "" {
			if cred, err := aws.CredentialsForRole(""); err == nil {
				config.awsAuth = cred
			} else {
				log.Fatal("Missing AWS_ACCESS_KEY or AWS_SECRET_KEY")
			}
		} else {
			config.awsAuth = keys
		}
	}

	if config.Region == "" {
		az, err := aws.GetMetadata(aws.MetadataAvailabilityZone)
		if err != nil {
			log.Fatalf("no region specified and failed to get from instance metadata: %+v", err)
		}
		config.Region = az[:len(az)-1]
	}

	config.ec2 = &ec2.EC2{
		Region: aws.Regions[config.Region],
		Client: &aws.Client{Auth: config.awsAuth},
	}

	if config.InstanceId == "" {
		var err error
		config.InstanceId, err = aws.GetMetadata(aws.MetadataInstanceID)
		if err != nil {
			log.Fatalf("Failed to get instance ID: %+v", err)
		}
	}

	// Lookup EBS volumes for devices
	if len(config.EBSVolumes) == 0 {
		vol, err := config.ec2.DescribeVolumes(nil, map[string][]string{
			"attachment.instance-id": []string{config.InstanceId},
		})
		if err != nil {
			log.Fatalf("Failed to get volumes: %+v", err)
		}
		config.EBSVolumes = make([]string, len(config.Devices))
		config.ebsVolumeInfo = make([]*ec2.Volume, len(config.Devices))
		count := len(config.Devices)
		for _, v := range vol {
			if v.Attachment != nil {
				for j, d := range config.Devices {
					if d == v.Attachment.Device {
						config.EBSVolumes[j] = v.VolumeId
						config.ebsVolumeInfo[j] = v
						count--
						break
					}
				}
			}
		}
		if count != 0 {
			log.Fatalf("Only found %d volumes out of an expected %d", len(config.Devices)-count, len(config.Devices))
		}
	} else {
		vol, err := config.ec2.DescribeVolumes(config.EBSVolumes, nil)
		if err != nil {
			log.Fatalf("Failed to get volumes: %+v", err)
		}
		if len(vol) != len(config.EBSVolumes) {
			log.Fatalf("Not all volumes found")
		}
		config.ebsVolumeInfo = make([]*ec2.Volume, len(config.EBSVolumes))
		for i, v := range vol {
			if config.EBSVolumes[i] != v.VolumeId {
				log.Fatalf("VolumeId mismatch")
			}
			config.ebsVolumeInfo[i] = v
		}
	}

	// MySQL config

	if config.Config != "" {
		if config.Config[0] == '~' {
			config.Config = os.Getenv("HOME") + config.Config[1:]
		}
		if err := readMySQLConfig(config.Config); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, path := range cnfSearchPath {
			if path[0] == '~' {
				path = os.Getenv("HOME") + path[1:]
			}
			if err := readMySQLConfig(path); err == nil {
				break
			}
		}
	}

	if config.Username == "" {
		config.Username = os.Getenv("MYSQL_USERNAME")
	}
	if config.Password == "" {
		config.Password = os.Getenv("MYSQL_PASSWORD")
	}

	enableTLS := config.CACert != "" && config.ClientCert != "" && config.ClientKey != ""
	if enableTLS {
		rootCertPool := x509.NewCertPool()
		pem, err := ioutil.ReadFile(config.CACert)
		if err != nil {
			log.Fatal(err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			log.Fatal("Failed to append PEM.")
		}
		clientCert := make([]tls.Certificate, 0, 1)
		certs, err := tls.LoadX509KeyPair(config.ClientCert, config.ClientKey)
		if err != nil {
			log.Fatal(err)
		}
		clientCert = append(clientCert, certs)
		mysql.RegisterTLSConfig("custom", &tls.Config{
			RootCAs:      rootCertPool,
			Certificates: clientCert,
		})
	}

	tlsOpt := ""
	if enableTLS {
		tlsOpt = "?tls=custom"
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s", config.Username, config.Password, config.Host, config.Port, "mysql", tlsOpt))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// db.SetMaxOpenConns(1)
	// db.SetMaxIdleConns(1)

	// test the connection to the database by running a ping against it
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	config.db = db

	if err := doIt(); err != nil {
		log.Fatal(err)
	}
}

func doIt() (err error) {
	var binlogName string
	var binlogPos int64
	if binlogName, binlogPos, err = lockDB(); err != nil {
		err = fmt.Errorf("Failed to lock DB: %s", err.Error())
		return
	}
	defer func() {
		e := unlockDB()
		// Don't overwrite other errors
		if err == nil {
			err = e
		} else if e != nil {
			log.Printf("Failed to unlock DB: %s", e.Error())
		}
	}()

	if err = freezeFS(); err != nil {
		err = fmt.Errorf("Failed to freeze filesystem: %s", err.Error())
		return
	}
	defer func() {
		e := thawFS()
		// Don't overwrite other errors
		if err == nil {
			err = e
		} else if e != nil {
			log.Printf("Failed to thaw filesystem: %s", e.Error())
		}
	}()

	err = snapshotEBS(binlogName, binlogPos)
	return
}

func lockDB() (string, int64, error) {
	fmt.Println("Locking database...")

	// Don't pass FLUSH TABLES statements on to replication slaves
	// as this can interfere with long-running queries on the slaves.
	if _, err := config.db.Exec("SET SQL_LOG_BIN=0"); err != nil {
		return "", 0, err
	}

	fmt.Println("Flushing tables without lock...")
	// Try a flush first without locking so the later flush with lock
	// goes faster.  This may not be needed as it seems to interfere with
	// some statements anyway.
	if _, err := config.db.Exec("FLUSH LOCAL TABLES"); err != nil {
		return "", 0, err
	}

	fmt.Println("Aquiring lock...")
	// Get a lock on the entire database
	if _, err := config.db.Exec("FLUSH LOCAL TABLES WITH READ LOCK"); err != nil {
		return "", 0, err
	}

	// This might be a slave database already
	// my $slave_status = $mysql_dbh->selectrow_hashref(q{ SHOW SLAVE STATUS });
	// $mysql_logfile           = $slave_status->{Slave_IO_State}
	//                          ? $slave_status->{Master_Log_File}
	//                          : undef;
	// $mysql_position          = $slave_status->{Read_Master_Log_Pos};
	// $mysql_binlog_do_db      = $slave_status->{Replicate_Do_DB};
	// $mysql_binlog_ignore_db  = $slave_status->{Replicate_Ignore_DB};

	fmt.Println("Getting master status...")
	// or this might be the master
	// File | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set
	var binlogFile, binlogDoDB, binlogIgnoreDB, executedGtidSet string
	var binlogPos int64
	if err := config.db.QueryRow("SHOW MASTER STATUS").Scan(&binlogFile, &binlogPos, &binlogDoDB, &binlogIgnoreDB, &executedGtidSet); err != nil {
		return "", 0, err
	}

	fmt.Printf("File=%s Position=%d Binlog_Do_DB=%s Binlog_Ignore_DB=%s Executed_Gtid_Set=%s\n", binlogFile, binlogPos, binlogDoDB, binlogIgnoreDB, executedGtidSet)

	if _, err := config.db.Exec("SET SQL_LOG_BIN=1"); err != nil {
		return binlogFile, binlogPos, err
	}

	return binlogFile, binlogPos, nil
}

func unlockDB() error {
	fmt.Println("Unlocking tables...")
	_, err := config.db.Exec("UNLOCK TABLES")
	return err
}

func freezeFS() error {
	fmt.Println("Freezing filesystem...")
	cmd := exec.Command(config.FreezeCmd, "-f", config.Filesystem)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func thawFS() error {
	fmt.Println("Thawing filesystem...")
	cmd := exec.Command(config.FreezeCmd, "-u", config.Filesystem)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func snapshotEBS(binlogName string, binlogPos int64) error {
	timestamp := time.Now().Format(time.RFC3339)
	for _, vol := range config.ebsVolumeInfo {
		fmt.Printf("Snapshotting %s (%s)...", vol.VolumeId, vol.Tags["Name"])
		res, err := config.ec2.CreateSnapshot(vol.VolumeId, fmt.Sprintf("%s %s", vol.Tags["Group"], timestamp))
		if err != nil {
			log.Fatalf("Failed to create snapshot of %s: %+v", vol.VolumeId, err)
		}
		fmt.Printf(" %s %s\n", res.SnapshotId, res.Status)
		tags := vol.Tags
		tags["BinlogName"] = binlogName
		tags["BinlogPos"] = strconv.FormatInt(binlogPos, 10)
		if err := config.ec2.CreateTags([]string{res.SnapshotId}, tags); err != nil {
			log.Printf("Failed to tag snapshot %s: %+v", res.SnapshotId, err)
		}
	}
	return nil
}
