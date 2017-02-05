package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"

	bp "github.com/cloudfoundry/libbuildpack"
)

func main() {
	bp_dir := os.Getenv("BUILDPACK_DIR")
	build_dir := os.Args[1]
	// cache_dir := os.Args[2]

	config, err := getConfig(build_dir)
	if err != nil {
		log.Fatal(err)
	}

	if len(config["root"]) == 0 {
		mvFilesToPublic(build_dir)
		config["root"] = "../public"
	}

	manifest, _ := bp.NewManifest(filepath.Join(bp_dir, "manifest.yml"))
	nginx, err := manifest.DefaultVersion("nginx")
	if err != nil {
		log.Fatal(err)
	}
	err = manifest.FetchDependency(nginx, "/tmp/nginx.tgz")
	if err != nil {
		log.Fatal(err)
	}
	err = bp.ExtractTarGz("/tmp/nginx.tgz", build_dir)
	if err != nil {
		log.Fatal(err)
	}
	// FIXME: Extract should do this
	err = os.Chmod(filepath.Join(build_dir, "nginx", "sbin", "nginx"), 0755)
	if err != nil {
		log.Fatal(err)
	}

	conf_dir := filepath.Join(build_dir, "nginx", "conf")
	err = os.MkdirAll(conf_dir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseGlob(filepath.Join(bp_dir, "conf", "nginx.conf")))
	fh, err := os.Create(filepath.Join(conf_dir, "nginx.conf.template"))
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(fh, config)
	if err != nil {
		log.Fatal(err)
	}

	err = Copy(filepath.Join(conf_dir, "mime.types"), filepath.Join(bp_dir, "conf", "mime.types"))
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(build_dir, "nginx", "sbin", "setup.sh"), []byte(`#!/usr/bin/env bash

mkfifo /tmp/nginx_access.log
cat < /tmp/nginx_access.log &
sed "s/__PORT__/$PORT/" < ./nginx/conf/nginx.conf.template > ./nginx/conf/nginx.conf
./nginx/sbin/nginx -p ./nginx/ -c conf/nginx.conf
	`), 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chmod(filepath.Join(build_dir, "nginx", "sbin", "setup.sh"), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func Copy(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func getConfig(build_dir string) (map[string]string, error) {
	yamlStr, err := ioutil.ReadFile(filepath.Join(build_dir, "Staticfile"))
	if err != nil {
		return nil, err
	}

	config := make(map[string]string)
	err = yaml.Unmarshal(yamlStr, &config)
	if len(config["root"]) > 0 {
		config["root"] = "../" + config["root"]
	}

	if _, err := os.Stat(filepath.Join(build_dir, "Staticfile.auth")); err == nil {
		bp.Log.BeginStep("Enabling basic authentication using Staticfile.auth")
		os.Rename(filepath.Join(build_dir, "Staticfile.auth"), filepath.Join(build_dir, "nginx", "conf", ".htpasswd"))
		config["auth_file"] = "conf/.htpasswd"
		bp.Log.Protip("Learn about basic authentication", "http://docs.cloudfoundry.org/buildpacks/staticfile/index.html#authentication")
	} else {
		delete(config, "auth_file")
	}

	if config["pushstate"] != "enabled" {
		delete(config, "pushstate")
	}
	if config["ssi"] != "enabled" {
		delete(config, "ssi")
	}
	if config["http_strict_transport_security"] != "true" {
		delete(config, "http_strict_transport_security")
	}
	if config["host_dot_files"] != "true" {
		delete(config, "host_dot_files")
	}

	return config, nil
}

func mvFilesToPublic(build_dir string) error {
	bp.Log.BeginStep("Copying project files into public/")

	filesToNotMove := map[string]bool{
		"Staticfile":      true,
		"Staticfile.auth": true,
		"manifest.yml":    true,
		"stackato.yml":    true,
		".profile":        true,
	}

	err := os.MkdirAll(filepath.Join(build_dir, "public"), 0755)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(build_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !filesToNotMove[file.Name()] {
			err := os.Rename(filepath.Join(build_dir, file.Name()), filepath.Join(build_dir, "public", file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
