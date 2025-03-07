package ui

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type rendezvousHostEnvTemplateData struct {
	RendezvousIP string
}

func saveRendezvousIPAddress(ipAddress string) error {
	_, err := os.Stat(RENDEZVOUS_HOST_ENV_PATH)
	if os.IsNotExist(err) {
		// TODO: Should we not expect RENDEZVOUS_HOST_ENV_PATH to always exist?
		// If so, this block can be removed.
		file, err := os.OpenFile(RENDEZVOUS_HOST_ENV_PATH, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not create and/or write to %s", RENDEZVOUS_HOST_ENV_PATH)
		}
		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("NODE_ZERO_IP=%s\nSERVICE_BASE_URL=http://%s:8090/\nIMAGE_SERVICE_BASE_URL=http://%s:8888/\n", ipAddress, ipAddress, ipAddress))
		if err != nil {
			return fmt.Errorf("error writing NODE_ZERO_IP to %s", RENDEZVOUS_HOST_ENV_PATH)
		}
	} else {
		templateData := &rendezvousHostEnvTemplateData{
			RendezvousIP: ipAddress,
		}
		err = templateRendezvousHostEnv(templateData)
		if err != nil {
			return err
		}
	}
	return nil
}

func templateRendezvousHostEnv(templateData interface{}) error {
	data, err := os.ReadFile(RENDEZVOUS_HOST_ENV_PATH)
	if err != nil {
		return err
	}

	tmpl := template.New(RENDEZVOUS_HOST_ENV_PATH).Funcs(template.FuncMap{"replace": replace})
	tmpl, err = tmpl.Parse(string(data))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, templateData); err != nil {
		panic(err)
	}
	templatedFileData := buf.Bytes()

	err = os.WriteFile(RENDEZVOUS_HOST_ENV_PATH, []byte(templatedFileData), 0644)
	if err != nil {
		return err
	}

	return nil
}

// replace is an utilitary function to do string replacement in templates.
func replace(input, from, to string) string {
	return strings.ReplaceAll(input, from, to)
}
