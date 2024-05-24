go install github.com/mitchellh/gox
gox -os="linux windows darwin" -arch="amd64" -output="build/{{.OS}}_{{.Arch}}/{{.Dir}}"
