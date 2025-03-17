package ui

import (
	"bytes"
	"os"
	"text/template"
)

type rendezvousHostEnvTemplateData struct {
	RendezvousIP string
}

func (u *UI) saveRendezvousIPAddress(ipAddress string) error {
	templateData := &rendezvousHostEnvTemplateData{
		RendezvousIP: ipAddress,
	}
	err := templateRendezvousHostEnv(templateData)
	if err != nil {
		return err
	}

	u.logger.Infof("Saved %s as Rendezvous IP", ipAddress)

	return nil
}

func templateRendezvousHostEnv(templateData interface{}) error {
	data, err := os.ReadFile(RENDEZVOUS_HOST_ENV_PATH)
	if err != nil {
		return err
	}

	tmpl := template.New(RENDEZVOUS_HOST_ENV_PATH)
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
