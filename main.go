package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func handleDownload(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var file, err = os.ReadFile("/go/media-server/ns_dump/" + vars["type"] + "/" + vars["key"])
	if err != nil {
		fmt.Println(err)
		res.Write([]byte("File Not Found."))
	}
	res.Write(file)
}

func ensureDir(dirName string) error {
	// fmt.Printf("Ensuring Directory: %s\n", dirName)
	err := os.Mkdir(dirName, 0666)
	if err != nil {
		info, errStat := os.Stat(dirName)
		if errStat != nil {
			return errStat
		}
		if !info.IsDir() {
			return errors.New("Path Exists and is not a directory.")
		}
		return nil
	}
	return nil
}

func handleUpload(res http.ResponseWriter, req *http.Request) {
	// handle form body
	req.ParseMultipartForm(10 << 20)
	file, _, err := req.FormFile("file")
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(res, "Error Uploading File.")

		return
	}
	defer file.Close()
	folderType := req.FormValue("type")
	fileName := req.FormValue("key")

	exists := ensureDir("/go/media-server/ns_dump/" + folderType + "/")
	if exists != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(res, "Error Uploading File.")
		return
	}
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(res, "Error Uploading File.")
		return
	}
	// fmt.Println("Writing File: ", "/go/media-server/ns_dump/"+folderType+"/"+fileName)
	err = os.WriteFile("/go/media-server/ns_dump/"+folderType+"/"+fileName, buf.Bytes(), 0666)
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(res, "Error Uploading File.")
		return
	}

	if err != nil {
		fmt.Println(err)
		fmt.Fprint(res, "Error Uploading File.")
		return
	}

	fmt.Fprint(res, "Upload Complete.")
	return
}

func handleRoot(res http.ResponseWriter, req *http.Request) {
	var foo = "Server is live."
	res.Write([]byte(foo))
}

func main() {
	fmt.Println("Starting Go Webserver - Live")
	server := mux.NewRouter()
	server.HandleFunc("/", handleRoot)
	server.HandleFunc("/m/{type}/{key}", handleDownload)
	server.HandleFunc("/upload", handleUpload)
	http.ListenAndServe(":6969", server)
}
