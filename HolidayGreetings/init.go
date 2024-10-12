package HolidayGreetings

import (
	_ "embed"
	"github.com/hduhelp/wechat_mp_server/hub"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

func init() {
	instance = &module{}
	hub.RegisterModule(instance)
}

var instance *module

type module struct {
	server *hub.Server
	cron   *cron.Cron
	wishes *AllWishes
	hub.UnimplementedModule
}

func (m *module) GetModuleInfo() hub.ModuleInfo {
	return hub.ModuleInfo{
		ID:       hub.NewModuleID("Necolo007", "HolidayWishes"),
		Instance: instance,
	}
}

type AllWishes struct {
	DBFWishes    []string `yaml:"DragonBoatFestivalWishes"`
	LDWishes     []string `yaml:"LaborDayWishes"`
	QMWishes     []string `yaml:"QMingFestivalWishes"`
	SpringWishes []string `yaml:"SpringFestivalWishes"`
	NDWishes     []string `yaml:"NationalDayWishes"`
}

//go:embed NationalDayWishes.yaml
var NationalDayWishes string

func (m *module) Init() {
	err := yaml.Unmarshal([]byte(NationalDayWishes), &m.wishes)
	if err != nil {
		return
	}
	m.cron = cron.New()
}
