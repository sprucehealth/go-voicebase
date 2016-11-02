package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/svc/deploy"
)

func prompt(scn *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	if !scn.Scan() {
		os.Exit(1)
	}
	return strings.TrimSpace(scn.Text())
}

func pprint(fs string, args ...interface{}) {
	fmt.Printf(fs, args...)
}

func printDeployableGroups(dgs []*deploy.DeployableGroup) {
	for _, dg := range dgs {
		printDeployableGroup(dg)
	}
}

func printDeployableGroup(dg *deploy.DeployableGroup) {
	pprint("Deployable Group: %s (name %s) (description %q)\n", dg.ID, dg.Name, dg.Description)
}

func printEnvironments(envs []*deploy.Environment) {
	for _, env := range envs {
		printEnvironment(env)
	}
}

func printEnvironment(env *deploy.Environment) {
	pprint("Environment: %s (name %s) (description %q) (deployable group %s) (prod %v)\n", env.ID, env.Name, env.Description, env.DeployableGroupID, env.IsProd)
}

func printDeployables(deps []*deploy.Deployable) {
	sort.Sort(deployableByName(deps))

	w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)
	fmt.Fprintf(w, "ID\tName\tDescription\tGitURL\n")
	for _, d := range deps {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.ID, d.Name, d.Description, d.GitURL)
	}
	if err := w.Flush(); err != nil {
		golog.Fatalf(err.Error())
	}
}

func printDeployable(dep *deploy.Deployable) {
	pprint("Deployable: %s (name %s) (description %q) (deployable group %s) (git url %s)\n", dep.ID, dep.Name, dep.Description, dep.DeployableGroupID, dep.GitURL)
}

func printEnvironmentConfigs(cs []*deploy.EnvironmentConfig) {
	for _, c := range cs {
		printEnvironmentConfig(c)
	}
}

func printEnvironmentConfig(c *deploy.EnvironmentConfig) {
	pprint("Environment Config: %s (environment %s) (status %q)\n", c.ID, c.EnvironmentID, c.Status)
	for k, v := range c.Values {
		pprint("\tName: %s Value: %s\n", k, v)
	}
}

func printDeployableConfigs(cs []*deploy.DeployableConfig) {
	for _, c := range cs {
		printDeployableConfig(c)
	}
}

func printDeployableConfig(c *deploy.DeployableConfig) {
	pprint("Deployable Config: %s (deployable %s) (environment %s) (status %q)\n", c.ID, c.DeployableID, c.EnvironmentID, c.Status)
	for _, kv := range valuesToSortedKV(c.Values) {
		pprint("\tName: %s Value: %s\n", kv[0], kv[1])
	}
}

func printDeployableVectors(vs []*deploy.DeployableVector) {
	for _, v := range vs {
		printDeployableVector(v)
	}
}

func printDeployableVector(v *deploy.DeployableVector) {
	sournceEnvironment := "none"
	if v.SourceType == deploy.DeployableVector_ENVIRONMENT_ID {
		sournceEnvironment = v.GetEnvironmentID()
	}
	pprint("Deployable Vector: %s (deployable %s) (source type %s) (source environment %s) (target environment %s)\n", v.ID, v.DeployableID, v.SourceType.String(), sournceEnvironment, v.TargetEnvironmentID)
}

func printDeployments(ds []*deploy.Deployment, depNames, envNames map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)
	fmt.Fprintf(w, "ID\tDeployableID\tEnvironmentID\tStatus\tDeployableConfigID\tType\tBuildNumber\n")
	for _, d := range ds {
		dep := depNames[d.DeployableID]
		if dep == "" {
			dep = d.DeployableID
		}
		env := envNames[d.EnvironmentID]
		if env == "" {
			env = d.EnvironmentID
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", d.ID, dep, env, d.Status, d.DeployableConfigID, d.Type, d.BuildNumber)
	}
	if err := w.Flush(); err != nil {
		golog.Fatalf(err.Error())
	}
}

func printDeployment(d *deploy.Deployment) {
	pprint("Deployment: %s (deployable %s) (environment %s) (status %s) (deployable config %s) (deployable vector %s) (type %s) (build number %s) (deployment number %d) (git hash %s)\n",
		d.ID, d.DeployableID, d.EnvironmentID, d.Status.String(), d.DeployableConfigID, d.DeployableVectorID, d.Type.String(), d.BuildNumber, d.DeploymentNumber, d.GitHash)
	switch d.Type {
	case deploy.Deployment_ECS:
		ecsDeployment := d.GetEcsDeployment()
		pprint("\tECS Deployment: (image %s) (cluster config name %s)\n", ecsDeployment.Image, ecsDeployment.ClusterDeployableConfigName)
	default:
		pprint("\tUnknown Deployment Type: %+v", d.DeploymentOneof)
	}
}

func valuesToSortedKV(v map[string]string) [][2]string {
	vals := make([][2]string, 0, len(v))
	for k, v := range v {
		vals = append(vals, [2]string{k, v})
	}
	sort.Sort(kv(vals))
	return vals
}

type deployableByName []*deploy.Deployable

func (d deployableByName) Len() int           { return len(d) }
func (d deployableByName) Swap(a, b int)      { d[a], d[b] = d[b], d[a] }
func (d deployableByName) Less(a, b int) bool { return d[a].Name < d[b].Name }

type kv [][2]string

func (kv kv) Len() int           { return len(kv) }
func (kv kv) Swap(a, b int)      { kv[a], kv[b] = kv[b], kv[a] }
func (kv kv) Less(a, b int) bool { return kv[a][0] < kv[b][0] }
