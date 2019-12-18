package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const imgEmbedFmt = `<html>
<head>
<style>
* {
margin: 0;
padding: 0;
}
.imgbox {
display: grid;
height: 100%%;
}
.center-fit {
max-width: 100%%;
max-height: 100vh;
margin: auto;
}
</style>
</head>
<body>
<div class="imgbox">
<img class="center-fit" src='%s'>
</div>
</body>
</html>`

//var globalTempDir = "/var/www/prettygood.dev/temp-images/"
var globalTempDir = "/tmp/temp-images/"

const (
	S3Region = "us-east-1"
	S3Bucket = "ok-zoomer-public-assets"
)

func UrlToUrl(sess *session.Session, inputImageUrl string) (string, error) {
	uploader := s3manager.NewUploader(sess)

	// download the image at inputImageUrl
	randomName := uuid.New().String()
	tempFile, err := ioutil.TempFile(globalTempDir, randomName + ".png")
	if err != nil {
		log.Fatalf("had trouble creating tempfile: %s", err.Error())
		return "", err
	}
	defer tempFile.Close()
	resp, err := http.Get(inputImageUrl)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("had trouble downloading image at %s: %s", inputImageUrl, err.Error())
		return "", err
	}
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		log.Fatalf("had trouble copying downloaded image to tempFile, err: %s", err.Error())
		return "", err
	}
	tempFile.Seek(0, io.SeekStart)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:                      tempFile,
		Bucket:                    aws.String(S3Bucket),
		Key:                       aws.String(fmt.Sprintf("/raw-images/%s.png", randomName)),
	})
	if err != nil {
		log.Fatalf("had trouble backing up input image to s3: err: %s", err.Error())
		return "", err
	}
	tempFile.Seek(0, io.SeekStart)

	// run the gif-making logic on the image
	outputPath := CreateGif(tempFile, 20)

	// upload the result to s3
	outputFile, err := os.Open(outputPath)
	if err != nil {
		log.Fatalf("had trouble opening the file at %s, err: %s", outputPath, err.Error())
		return "", err
	}
	defer outputFile.Close()
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:                      outputFile,
		Bucket:                    aws.String(S3Bucket),
		Key:                       aws.String(fmt.Sprintf("/gifs/%s.gif", randomName)),
	})

	// return the url to the gif object on s3
	return result.Location, nil

}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println("Error parsing multipart form")
		fmt.Println(err)
		return
	}

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile(globalTempDir, "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	tempFile.Seek(0, io.SeekStart)

	// create the dang gif
	outputPath := CreateGif(tempFile, 20)
	// return that we have successfully uploaded our file!
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, imgEmbedFmt, outputPath)
}

func setupRoutes() {
	awsSess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(S3Region),
	}))
	http.HandleFunc("/sms", GetTwilioHandler(awsSess))
	http.HandleFunc("/upload", uploadFile)
	//http.Handle("/temp-images/", http.StripPrefix("/temp-images/", http.FileServer(http.Dir("temp-images"))))
	//http.Handle(globalTempDir,
	//	http.StripPrefix(globalTempDir,
	//		http.FileServer(http.Dir(globalTempDir))))
	http.ListenAndServe(":8080", nil)
	fmt.Println("successfully set up routes!")
}

//func removeAndLog(pathToRemove string) {
//	err := os.RemoveAll(pathToRemove)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("successfully removed " + pathToRemove)
//}
//
func main() {
	fmt.Println("Hello World")
	//curdir, err := os.Getwd()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//var globalTempDirAbsPath = ""
	//globalTempDirAbsPath, err = ioutil.TempDir(curdir, "example")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer removeAndLog(globalTempDirAbsPath)

	setupRoutes()
}
