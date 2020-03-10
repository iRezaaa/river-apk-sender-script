package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var sendReleases = false
	var sendDebugs = false
	var sendEnterprises = false

	for _, arg := range os.Args {
		if arg == "-r" {
			sendReleases = true
		} else if arg == "-e" {
			sendEnterprises = true
		} else if arg == "-d" {
			sendDebugs = true
		}
	}

	if !sendEnterprises && !sendDebugs && !sendReleases {
		panic("you must choose at least one build type")
	}

	remoteURL := "https://drive.ronaksoft.com/admin/files"

	client := &http.Client{}

	apksPath, err := os.Open("app/build/outputs/apk")

	if err != nil {
		panic("Apk folder not found! in 'app/build/outputs/apk'")
	}

	var debugPath, releasesPath, enterprisesPath string

	buildTypesPaths, err := ioutil.ReadDir(apksPath.Name())

	if err != nil {
		panic(err)
	}

	for _, f := range buildTypesPaths {
		if f.Name() == "debug" {
			debugPath = apksPath.Name() + "/" + f.Name()
		} else if f.Name() == "release" {
			releasesPath = apksPath.Name() + "/" + f.Name()
		} else if f.Name() == "enterprise" {
			enterprisesPath = apksPath.Name() + "/" + f.Name()
		}
	}

	var releases []string
	var debugs []string
	var enterprises []string

	i := 0
	for i < 3 {
		var currentPath string
		var currentName string

		if i == 0 {
			if len(debugPath) == 0 {
				i++
				continue
			}
			currentPath = debugPath
			currentName = "debug"
		} else if i == 1 {
			if len(releasesPath) == 0 {
				i++
				continue
			}
			currentPath = releasesPath
			currentName = "release"
		} else if i == 2 {
			if len(enterprisesPath) == 0 {
				i++
				continue
			}
			currentPath = enterprisesPath
			currentName = "enterprise"
		}

		var lastArmV8 string
		var lastArmV7 string
		var lastX86 string
		var lastX86_64 string

		err = filepath.Walk(currentPath, func(path string, f os.FileInfo, err error) error {
			if strings.HasPrefix(path, currentPath) && strings.HasSuffix(f.Name(), "arm64-v8a-"+currentName+".apk") {
				lastArmV8 = path
			} else if strings.HasPrefix(path, currentPath) && strings.HasSuffix(f.Name(), "armeabi-v7a-"+currentName+".apk") {
				lastArmV7 = path
			} else if strings.HasPrefix(path, currentPath) && strings.HasSuffix(f.Name(), "x86_64-"+currentName+".apk") {
				lastX86_64 = path
			}else if strings.HasPrefix(path, currentPath) && strings.HasSuffix(f.Name(), "x86-"+currentName+".apk") {
				lastX86 = path
			}

			return nil
		})

		if len(lastArmV8) != 0 {
			switch i {
			case 0 :
				debugs = append(debugs, lastArmV8)
			case 1 :
				releases = append(releases,lastArmV8)
			case 2 :
				enterprises = append(enterprises,lastArmV8)
			}
		}

		if len(lastArmV7) != 0 {
			switch i {
			case 0 :
				debugs = append(debugs, lastArmV7)
			case 1 :
				releases = append(releases,lastArmV7)
			case 2 :
				enterprises = append(enterprises,lastArmV7)
			}
		}

		if len(lastX86) != 0 {
			switch i {
			case 0 :
				debugs = append(debugs, lastX86)
			case 1 :
				releases = append(releases,lastX86)
			case 2 :
				enterprises = append(enterprises,lastX86)
			}
		}

		if len(lastX86_64) != 0 {
			switch i {
			case 0 :
				debugs = append(debugs, lastX86_64)
			case 1 :
				releases = append(releases,lastX86_64)
			case 2 :
				enterprises = append(enterprises,lastX86_64)
			}
		}

		if err != nil {
			panic(err)
		}

		i++
	}

	if sendReleases {
		print("Uploading Releases started...")
		for _, path := range releases {
			print("\n" + "Uploading " + path + " ...")

			file, err := os.Open(path)

			if err != nil {
				panic(err)
			}

			values := map[string]io.Reader{
				"files[]":  file, // lets assume its this file
				"path":     strings.NewReader("/Releases/River/Android/Private/"),
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
		for _, path := range debugs {
			print("\n" + "Uploading " + path + " ...")

			file, err := os.Open(path)

			if err != nil {
				panic(err)
			}

			values := map[string]io.Reader{
				"files[]":  file, // lets assume its this file
				"path":     strings.NewReader("/Releases/River/Android/Debugs/"),
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
		for _, path := range enterprises {
			print("\n" + "Uploading " + path + " ...")

			file, err := os.Open(path)

			if err != nil {
				panic(err)
			}

			values := map[string]io.Reader{
				"files[]":  file, // lets assume its this file
				"path":     strings.NewReader("/Releases/River/Android/Enterprises/"),
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
