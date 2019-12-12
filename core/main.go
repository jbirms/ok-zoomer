package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

var globalTempDir = "/tmp/temp-images/"

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
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
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
	outputPath := CreateGif(tempFile, 16)
	// return that we have successfully uploaded our file!
	//fmt.Fprintf(w, "Successfully Uploaded File at %s\n", outputPath)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//fmt.Fprintf(w, "<img src=\"%s\" alt=\"test\">", outputPath)

	fmt.Fprintf(w, imgEmbedFmt, outputPath)
}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	//http.Handle("/temp-images/", http.StripPrefix("/temp-images/", http.FileServer(http.Dir("temp-images"))))
	http.Handle(globalTempDir,
		http.StripPrefix(globalTempDir,
			http.FileServer(http.Dir(globalTempDir))))
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
