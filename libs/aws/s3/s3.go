package s3

// TODO: retries

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sprucehealth/backend/libs/aws"
)

type S3 struct {
	aws.Region
	Client *aws.Client
}

func (s3 *S3) buildPath(bucket, path string) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}
	return "/" + bucket + path
}

func (s3 *S3) buildURL(bucket, path string) string {
	return s3.S3Endpoint + s3.buildPath(bucket, path)
}

func (s3 *S3) Do(req *http.Request, result interface{}) (*http.Response, error) {
	if s3.Client.HTTPClient == nil {
		s3.Client.HTTPClient = http.DefaultClient
	}
	s3.sign(req)
	res, err := s3.Client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		defer res.Body.Close()
		return res, ParseErrorResponse(res)
	}
	if result != nil {
		defer res.Body.Close()
		return res, xml.NewDecoder(res.Body).Decode(result)
	}
	return res, nil
}

// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectGET.html
func (s3 *S3) Get(bucket, path string) ([]byte, error) {
	rd, err := s3.GetReader(bucket, path)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	return ioutil.ReadAll(rd)
}

// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectGET.html
func (s3 *S3) GetReader(bucket, path string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", s3.buildURL(bucket, path), nil)
	if err != nil {
		return nil, err
	}
	res, err := s3.Do(req, nil)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectHEAD.html
func (s3 *S3) Head(bucket, path string) (http.Header, error) {
	req, err := http.NewRequest("HEAD", s3.buildURL(bucket, path), nil)
	if err != nil {
		return nil, err
	}
	res, err := s3.Do(req, nil)
	if err != nil {
		return nil, err
	}
	if res.Body != nil {
		res.Body.Close()
	}
	return res.Header, nil
}

// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPUT.html
func (s3 *S3) Put(bucket, path string, data []byte, contType string, perm ACL, additionalHeaders map[string][]string) error {
	h := md5.New()
	h.Write(data)
	md5Sum := base64Std.EncodeToString(h.Sum(nil))

	dataReader := bytes.NewReader(data)
	req, err := http.NewRequest("PUT", s3.buildURL(bucket, path), dataReader)
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(data))
	req.Header.Set("Content-MD5", md5Sum)
	if contType != "" {
		req.Header.Set("Content-Type", contType)
	}
	if perm != "" {
		req.Header.Set(HeaderACL, string(perm))
	}
	for key, values := range additionalHeaders {
		req.Header[key] = values
	}
	res, err := s3.Do(req, nil)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPUT.html
func (s3 *S3) PutFrom(bucket, path string, rd io.Reader, size int64, contType string, perm ACL, additionalHeaders map[string][]string) error {
	req, err := http.NewRequest("PUT", s3.buildURL(bucket, path), rd)
	if err != nil {
		return err
	}
	req.ContentLength = size
	if contType != "" {
		req.Header.Set("Content-Type", contType)
	}
	if perm != "" {
		req.Header.Set(HeaderACL, string(perm))
	}
	for key, values := range additionalHeaders {
		req.Header[key] = values
	}
	res, err := s3.Do(req, nil)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectDELETE.html
func (s3 *S3) Delete(bucket, path string) error {
	req, err := http.NewRequest("DELETE", s3.buildURL(bucket, path), nil)
	if err != nil {
		return err
	}
	res, err := s3.Do(req, nil)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

type ListBucketParams struct {
	// A delimiter is a character you use to group keys.
	// All keys that contain the same string between the prefix, if specified, and the first occurrence of
	// the delimiter after the prefix are grouped under a single result element, CommonPrefixes. If you don't
	// specify the prefix parameter, then the substring starts at the beginning of the key. The keys that are
	// grouped under CommonPrefixes result element are not returned elsewhere in the response.
	Delimiter string
	// Market specifies the key to start with when listing objects in a bucket. Amazon S3 returns object keys in
	// alphabetical order, starting with key after the marker in order.
	Marker string
	// MaxKeys sets the maximum number of keys returned in the response body. You can add this to your request
	// if you want to retrieve fewer than the default 1000 keys.
	// The response might contain fewer keys but will never contain more. If there are additional keys that
	// satisfy the search criteria but were not returned because max-keys was exceeded, the response contains
	// <IsTruncated>true</IsTruncated>. To return the additional keys, see marker.
	MaxKeys int
	// Prefix limits the response to keys that begin with the specified prefix. You can use prefixes to separate
	// a bucket into different groupings of keys. (You can think of using prefix to make groups in the same way
	// you'd use a folder in a file system.)
	Prefix string
}

// ListBucket returns some or all (up to 1000) of the objects in a bucket..
// To use this implementation of the operation, you must have READ access to the bucket.
// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketGET.html
func (s3 *S3) ListBucket(bucket string, params *ListBucketParams) (*ListBucketsResult, error) {
	p := url.Values{}
	if params != nil {
		if params.Delimiter != "" {
			p.Set("delimiter", params.Delimiter)
		}
		if params.Marker != "" {
			p.Set("marker", params.Marker)
		}
		if params.MaxKeys > 0 {
			p.Set("max-keys", strconv.Itoa(params.MaxKeys))
		}
		if params.Prefix != "" {
			p.Set("prefix", params.Prefix)
		}
	}

	req, err := http.NewRequest("GET", s3.buildURL(bucket, "")+"?"+p.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var res ListBucketsResult
	if _, err := s3.Do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
