package docker

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/agurinov/dnskek/log"
)

const (
	successLogStatus = "SUCCESS"
	failLogStatus    = "FAIL"
)

var (
	localDrivers = map[string]bool{
		"virtualbox": true,
	}

	validStates = map[string]bool{
		"Running": true,
	}
)

type Registry struct {
	items []*Machine
}

func NewRegistry() *Registry {
	// create new empty registry
	reg := new(Registry)
	// fill registry with all available docker machines
	reg.fill()
	// provide ready registry
	return reg
}

func (reg *Registry) fill() {
	defer log.Debugf("Registry.fill() -> %s", successLogStatus)
	// prepare command to fetch docker mahines
	cmd, err := exec.LookPath("docker-machine")
	if err != nil {
		panic(err)
	} // cannot find executable in PATH
	args := []string{"ls", "--format", "{{.Name}}|{{.DriverName}}|{{.State}}|{{.URL}}|{{.DockerVersion}}"}
	// get docker machines (command output)
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		panic(err)
	} // cannot get docker machines

	// parse output
	output := strings.Trim(string(out), "\n")
	// iterate over raw data and fill registry Machine instances
	for _, machine_row := range strings.Split(output, "\n") {
		machine_data := strings.Split(machine_row, "|")
		// parse data and get machine struct
		dm := newMachine(machine_data[0], machine_data[1], machine_data[2], machine_data[3], machine_data[4])
		// register machine
		reg.items = append(reg.items, dm)
		defer log.Debugf("Registry.fill.Machine(name=%q) -> %s", dm.Name, successLogStatus)
	}
}

func (reg *Registry) update() {
	defer log.Debugf("Registry.update() -> %s", successLogStatus)
	// clear registry
	reg.items = nil
	// fill again
	reg.fill()
}

func (reg *Registry) ResolveMachineByName(name string) (*Machine, error) {
	// debug logging
	defer func() {
		// if err != nil {
		//     status := failLogStatus
		// } else {
		//     status := successLogStatus
		// }
		log.Debugf("Registry.ResolveMachineByName(name=%q) -> %s", name, successLogStatus)
	}()
	// iterate over registry
	log.Debug("machines", reg.items)
	for _, dm := range reg.items {
		// compile regexp for resolving
		p, err := regexp.Compile("^" + subdomainsRegexExpression + dm.DnsName() + "$")
		
		log.Debug(p.String())
		
		
		// check machine for conditions
		switch {
		case err != nil, !p.MatchString(name): // regex not valid or no match
			log.Debug("regex not valid or no match")
			continue
		case !validStates[dm.State]: // machine found but not Running
			log.Debug("err invalid state", dm.State)
			return nil, ErrMachineNotRunning
		case dm.IP == nil: // no IP
			log.Debug("err no ip")
			return nil, ErrMachineNoIP
		case !localDrivers[dm.DriverName]: // driver not for local usage
			log.Debug("err not local driver", dm.DriverName)
			return nil, ErrMachineNotLocalDriver
		default:
			log.Debug("all wright, ip", dm.IP)
			return dm, nil // this is valid machine
		}
	}
	log.Debug("not matches!!!!")
	// no matches -> return error
	return nil, ErrMachineNotExist
}
