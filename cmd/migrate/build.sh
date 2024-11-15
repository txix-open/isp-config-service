go install github.com/mitchellh/gox@latest
gox -os="linux windows darwin" -arch="amd64" -output="build/{{.OS}}_{{.Arch}}/{{.Dir}}"
