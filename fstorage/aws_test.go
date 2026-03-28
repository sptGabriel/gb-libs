package fstorage_test

import (
	"bytes"
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/sptGabriel/defaults/fstorage"
)

const defaultRegion = "us-east-1"

func (s *fileStorageTestSuite) TestShouldUploadFileCorrectly() {
	require := s.Require()

	motoInstance := s.instance
	tLogger := s.logger

	// setup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	s3Url, err := motoInstance.ConnectionString(ctx)
	require.NoError(err)

	creds := fstorage.AWSS3Credentials{
		AccessKey: "test",
		SecretKey: "test",
		URL:       s3Url,
		Region:    defaultRegion,
	}

	client, err := fstorage.NewAWSS3FileStorage(ctx, creds)
	require.NoError(err, tLogger.String())

	contents := make([]byte, 10*1024*1024)
	contents[0] = 'S'
	contents[1] = 'P'
	contents[2] = 'D'
	contents[3] = 'F'
	contents[4] = '-'

	bucketName := "aws-s3-test-bucket"
	filename := "A simple file for upload.pdf"
	path := "storage/" + filename

	// act - check bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	require.NoError(err, tLogger.String())
	require.False(exists, "bucket must not exist")

	// act - create bucket
	err = client.CreateBucket(ctx, creds.Region, bucketName)
	require.NoError(err, tLogger.String())

	// act - check bucket exists
	exists, err = client.BucketExists(ctx, bucketName)
	require.NoError(err, tLogger.String())
	require.True(exists, "bucket must exist")

	// act - upload
	ext := filepath.Ext(filename)
	metadata := map[string]string{
		"name":      filename,
		"extension": ext,
		"mimetype":  mime.TypeByExtension(ext),
	}

	err = client.UploadFile(
		ctx,
		bucketName,
		path,
		bytes.NewBuffer(contents),
		metadata,
	)
	require.NoError(err, tLogger.String())

	// act - get temporary downloadURL
	downloadURL, err := client.GetDownloadPresignedURL(
		ctx,
		bucketName,
		path,
		5*time.Minute,
	)

	// assert
	require.NoError(err, tLogger.String())
	require.NotEmpty(downloadURL, tLogger.String())

	_, err = url.Parse(downloadURL)
	require.NoError(err, tLogger.String())

	// assert by download file to disk
	dir := s.T().TempDir()
	downloadPath := filepath.Join(dir, filename)
	out, err := os.Create(downloadPath)
	require.NoError(err, tLogger.String())
	defer func() {
		err := out.Close()
		require.NoError(err, tLogger.String())
	}()

	resp, err := http.Get(downloadURL)
	require.NoError(err, tLogger.String())
	defer func() {
		err := resp.Body.Close()
		require.NoError(err, tLogger.String())
	}()

	// write the body to file
	n, err := io.Copy(out, resp.Body)
	require.NoError(err, tLogger.String())
	require.Equal(int64(len(contents)), n)
}
