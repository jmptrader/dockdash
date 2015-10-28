package view

import (
	docklistener "github.com/byrnedo/dockdash/docklistener"
	. "github.com/byrnedo/dockdash/logger"
	goDocker "github.com/fsouza/go-dockerclient"
	ui "github.com/gizak/termui"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DockerInfoType int

const (
	ImageInfo DockerInfoType = iota
	PortInfo
	BindInfo
	CommandInfo
	EntrypointInfo
	EnvInfo
	VolumesInfo
	TimeInfo
)

var InfoHeaders map[DockerInfoType]string = map[DockerInfoType]string{
	ImageInfo:      "Image",
	PortInfo:       "Ports",
	BindInfo:       "Mounts",
	CommandInfo:    "Command",
	EntrypointInfo: "Entrypoint",
	EnvInfo:        "Envs",
	VolumesInfo:    "Volumes",
	TimeInfo:       "Created At",
}

const MaxContainers = 1000
const MaxHorizPosition = int(TimeInfo)

type View struct {
	Header   *ui.Par
	CpuChart *ui.BarChart
	MemChart *ui.BarChart
	NameList *ui.List
	InfoList *ui.List
}

func createContainerList() *ui.List {
	list := ui.NewList()
	list.ItemFgColor = ui.ColorCyan
	list.Border.FgColor = ui.ColorBlack
	list.HasBorder = true
	return list
}

type ContainerSlice []*goDocker.Container

func (p ContainerSlice) Len() int {
	return len(p)
}

func (p ContainerSlice) Less(i, j int) bool {
	return p[i].State.StartedAt.After(p[j].State.StartedAt)
}

func (p ContainerSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func createDockerLineChart() *ui.LineChart {
	lc := ui.NewLineChart()
	lc.Border.Label = "Container Numbers"
	lc.Height = 10
	lc.AxesColor = ui.ColorWhite
	lc.LineColor = ui.ColorRed | ui.AttrBold
	lc.Mode = "line"
	return lc
}

func NewView() *View {

	var view = View{}

	view.Header = ui.NewPar("Containers")
	view.Header.HasBorder = false
	view.Header.Text = " Dockdash - Interactive realtime container inspector"
	view.Header.Height = 3

	view.NameList = createContainerList()
	view.NameList.Border.Label = "Name"

	view.InfoList = createContainerList()
	view.InfoList.Border.Label = "Image"

	view.CpuChart = ui.NewBarChart()
	view.CpuChart.HasBorder = true
	view.CpuChart.Border.Label = "%CPU"
	view.CpuChart.Border.FgColor = ui.ColorBlack
	view.CpuChart.Height = 8

	view.MemChart = ui.NewBarChart()
	view.MemChart.HasBorder = true
	view.MemChart.Border.Label = "%MEM"
	view.MemChart.Border.FgColor = ui.ColorBlack
	view.MemChart.Height = 8
	return &view
}

func (v *View) SetLayout() {
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, v.Header),
		),
		ui.NewRow(
			ui.NewCol(12, 0, v.CpuChart),
		),
		ui.NewRow(
			ui.NewCol(12, 0, v.MemChart),
		),
		ui.NewRow(
			ui.NewCol(3, 0, v.NameList),
			ui.NewCol(9, 0, v.InfoList),
		),
	)
}

func (v *View) Align() {
	ui.Body.Align()
}

func (v *View) ResetSize() {
	ui.Body.Width = ui.TermWidth()
	ui.Body.Align()
}

func (v *View) Render() {
	ui.Render(ui.Body)
}

func (v *View) RenderStats(statsCharts *docklistener.StatsMsg, offset int) {
	v.CpuChart.Data = statsCharts.CpuChart.Data[offset:]
	v.CpuChart.DataLabels = statsCharts.CpuChart.DataLabels[offset:]
	v.MemChart.Data = statsCharts.MemChart.Data[offset:]
	v.MemChart.DataLabels = statsCharts.MemChart.DataLabels[offset:]
	v.Render()
}

func (v *View) RenderContainers(containers map[string]*goDocker.Container, infoType DockerInfoType, listOffset int) {
	names, info := getNameAndInfoOfContainers(containers, listOffset, infoType)
	var height = len(names) + 2
	v.NameList.Height = height
	v.NameList.Items = names
	v.InfoList.Height = height
	v.InfoList.Items = info
	v.InfoList.Border.Label = InfoHeaders[infoType]
	v.Render()
}

func getNameAndInfoOfContainers(containers map[string]*goDocker.Container, offset int, infoType DockerInfoType) ([]string, []string) {
	var numContainers = len(containers)
	if offset > numContainers {
		offset = numContainers - 1
	}

	var (
		numContainersSubset = numContainers - offset
		names               = make([]string, numContainersSubset)
		info                = make([]string, numContainersSubset)
		containersSorted    = mapValuesSorted(containers)
	)
	for index, cont := range containersSorted {
		if index < offset {
			continue
		}

		var containerNumber = numContainers - index

		names[index-offset] = "(" + strconv.Itoa(containerNumber) + ") " + cont.ID[:12] + " " + strings.TrimLeft(cont.Name, "/")
		switch infoType {
		case ImageInfo:
			info[index-offset] = cont.Config.Image
		case PortInfo:
			info[index-offset] = createPortsString(cont.NetworkSettings.Ports)
		case BindInfo:
			info[index-offset] = strings.TrimRight(strings.Join(cont.HostConfig.Binds, ","), ",")
		case CommandInfo:
			info[index-offset] = cont.Path + " " + strings.Join(cont.Args, " ")
		case EnvInfo:
			info[index-offset] = strings.TrimRight(strings.Join(cont.Config.Env, ","), ",")
		case EntrypointInfo:
			info[index-offset] = strings.Join(cont.Config.Entrypoint, " ")
		case VolumesInfo:
			volStr := ""
			for intVol, hostVol := range cont.Volumes {
				volStr += intVol + ":" + hostVol + ","
			}
			info[index-offset] = strings.TrimRight(volStr, ",")
		case TimeInfo:
			info[index-offset] = cont.State.StartedAt.Format(time.RubyDate)
		default:
			Error.Println("Unhandled info type", infoType)
		}
	}
	return names, info
}

func mapValuesSorted(mapToSort map[string]*goDocker.Container) (sorted ContainerSlice) {

	sorted = make(ContainerSlice, len(mapToSort))
	var i = 0
	for _, val := range mapToSort {
		sorted[i] = val
		i++
	}
	sort.Sort(sorted)
	return
}

func createPortsString(ports map[goDocker.Port][]goDocker.PortBinding) (portsStr string) {

	for intPort, extHostPortList := range ports {
		if len(extHostPortList) == 0 {
			portsStr += intPort.Port() + "->N/A,"
		}
		for _, extHostPort := range extHostPortList {
			portsStr += intPort.Port() + "->" + extHostPort.HostIP + ":" + extHostPort.HostPort + ","
		}
	}
	return strings.TrimRight(portsStr, ",")
}
