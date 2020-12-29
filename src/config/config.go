package config

type PropertyHolder struct {
	Bind           string   `cfg:"bind"`
	Port           int      `cfg:"port"`
	AppendOnly     bool     `cfg:"appendOnly"`
	AppendFIlename string   `cfg:"appendFilename"`
	MaxClients     int      `cfg:"maxclients"`
	Peers          []string `cfg:"peers"`
	Self           string   `cfg:"self"`
}

var Properties *PropertyHolder

func init(){
	// default config
	Properties = &PropertyHolder{
		Bind: "127.0.0.1",
		Port: 6383,
		AppendOnly: false,
	}
}