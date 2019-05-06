package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
			"os"
	"strings"
	"io/ioutil"
	"path/filepath"
	)

func main() {
	var sendReleases = false
	var sendDebugs = false
	var sendEnterprises = false

	for _ , arg := range os.Args {
		if arg == "-r" {
			sendReleases = true
		}else if arg == "-e" {
			sendEnterprises = true
		}else if arg == "-d" {
			sendDebugs = true
		}
	}

	if !sendEnterprises && !sendDebugs && !sendReleases {
		panic("you must choose at least one build type")
	}

	remoteURL := "https://drive.ronaksoft.com/admin/files"

	client := &http.Client{}

	apksPath , err := os.Open("app/build/outputs/apk")

	if err != nil {
		panic("Apk folder not found! in 'app/build/outputs/apk'")
	}


	var arm64Path,armv7Path,x64Path,x86Path string

	archs , err := ioutil.ReadDir(apksPath.Name())

	if err != nil {
		panic(err)
	}

	for _, f := range archs {
		if f.Name() == "arm64" {
			arm64Path = apksPath.Name() + "/" + f.Name()
		}else if f.Name() == "armv7" {
			armv7Path = apksPath.Name() + "/" + f.Name()
		}else if f.Name() == "x64" {
			x64Path = apksPath.Name() + "/" + f.Name()
		}else if f.Name() == "x86" {
			x86Path = apksPath.Name() + "/" + f.Name()
		}
	}

	var releases []string
	var debugs []string
	var enterprises []string

	i := 0
	for i <= 4 {
		var currentPath string

		if i == 0 {
			currentPath = arm64Path
			//name = "Arm64"
		}else if i == 1 {
			currentPath = armv7Path
			//name = "Armv7"
		}else if i == 2 {
			currentPath = x64Path
			//name = "X64"
		}else if i == 3 {
			currentPath = x86Path
			//name = "X86"
		}

		var lastRelease string
		var lastDebug string
		var lastEnterprise string

		err = filepath.Walk(currentPath, func (path string, f os.FileInfo, err error) error {
			if strings.HasPrefix(path,currentPath + "/release") && strings.HasSuffix(f.Name(), "apk") {
				lastRelease = path
			}else if strings.HasPrefix(path,currentPath + "/debug") && strings.HasSuffix(f.Name(), "apk") {
				lastDebug = path
			}else if strings.HasPrefix(path,currentPath + "/enterprise") && strings.HasSuffix(f.Name(), "apk") {
				lastEnterprise = path
			}

			return nil
		})

		if len(lastRelease) != 0 {
			releases = append(releases, lastRelease)
		}

		if len(lastDebug) != 0 {
			debugs = append(debugs, lastDebug)
		}

		if len(lastEnterprise) != 0 {
			enterprises = append(enterprises, lastEnterprise)
		}

		if err != nil {
			panic(err)
		}


		i = i + 1
	}



	if sendReleases {
		print("Uploading Releases started...")
		for _ , path := range releases {
			print("\n" + "Uploading " + path + " ...")

			file , err := os.Open(path)

			if err != nil {
				panic(err)
			}

			values := map[string]io.Reader{
				"files[]":  file, // lets assume its this file
				"path": strings.NewReader("/Releases/River/Android/Releases/"),
				"username": strings.NewReader("ireza"),
				"password": strings.NewReader("21506426"),
			}

			err = Upload(client, remoteURL, values)

			if err != nil {
				panic(err)
			}

			print("\n" + path + "Uploaded!")
		}
	}

	if sendDebugs {
		print("\n" + "Uploading Debugs started...")
		for _ , path := range debugs {
			print("\n" + "Uploading " + path + " ...")

			file , err := os.Open(path)

			if err != nil {
				panic(err)
			}

			values := map[string]io.Reader{
				"files[]":  file, // lets assume its this file
				"path": strings.NewReader("/Releases/River/Android/Debugs/"),
				"username": strings.NewReader("ireza"),
				"password": strings.NewReader("21506426"),
			}

			err = Upload(client, remoteURL, values)

			if err != nil {
				panic(err)
			}

			print("\n" + path + "Uploaded!")
		}
	}

	if sendEnterprises {
		print("\n" + "Uploading Enterprises started...")
		for _ , path := range enterprises {
			print("\n" + "Uploading " + path + " ...")

			file , err := os.Open(path)

			if err != nil {
				panic(err)
			}

			values := map[string]io.Reader{
				"files[]":  file, // lets assume its this file
				"path": strings.NewReader("/Releases/River/Android/Enterprises/"),
				"username": strings.NewReader("ireza"),
				"password": strings.NewReader("21506426"),
			}

			err = Upload(client, remoteURL, values)

			if err != nil {
				panic(err)
			}

			print("\n" + path + "Uploaded!")
		}
	}
}

func Upload(client *http.Client, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

