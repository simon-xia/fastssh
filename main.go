package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/simon-xia/fzf/src"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	defaultConfCnt  = 10
	defaultConffile = ".fastsshrc"
	defaultScript   = ".fast_login.sh"
	confSpliter     = "|"
	confFieldCnt    = 6
	headerLine      = 2
	conffileFormat  = "name|host|user|password|port|comment"

	expectScriptTmpl = `set timeout 30
trap {
    set rows [stty rows]
    set cols [stty columns]
    stty rows $rows columns $cols < $spawn_out(slave,name)
} WINCH
spawn ssh -p%s -l %s %s
expect "password:"
send   %s\r
interact`

	bar = ``
)

var (
	conffile = flag.String("f", defaultConffile, "config file name, under home directory")

	defaultScriptPath = ""
)

type LoginInfo struct {
	Name     string
	Addr     string
	User     string
	Password string
	Port     string
	Comment  string
}

func (info *LoginInfo) expectScript() string {
	return fmt.Sprintf(expectScriptTmpl, info.Port, info.User, info.Addr, info.Password)
}

func genScript(info LoginInfo) {
	f, err := os.Create(defaultScriptPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprint(w, info.expectScript())
	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

func parseConf(line string) LoginInfo {
	fileds := strings.Split(line, confSpliter)
	if len(fileds) != confFieldCnt {
		log.Fatalf("parse conf file failed, line(%s), format is %s", line, conffileFormat)
	}

	return LoginInfo{
		Name:     fileds[0],
		Addr:     fileds[1],
		User:     fileds[2],
		Password: fileds[3],
		Port:     fileds[4],
		Comment:  fileds[5],
	}
}

func loadLoginInfoConf(conffile string) []LoginInfo {

	f, err := os.OpenFile(conffile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open config file(%s) error: %v", conffile, err)
	}
	defer f.Close()

	loginInfos := make([]LoginInfo, 0, defaultConfCnt)

	sc := bufio.NewScanner(f)
	header := true
	for sc.Scan() {
		if (header) { // skip header
			header = false
			continue
		}
		loginInfos = append(loginInfos, parseConf(sc.Text()))
	}

	if err := sc.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
	}

	return loginInfos
}

func renderTable(infos []LoginInfo, w io.Writer) {

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"No", "Name", "IP/Hostname", "User", "Comment", "Port"})
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	for i, info := range infos {
		table.Append([]string{strconv.Itoa(i + 1), info.Name, info.Addr, info.User, info.Comment, info.Port})
	}

	table.Render()
}

func getExpectPath() string {
	out, err := exec.Command("which", "expect").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func runSSH(info LoginInfo) {

	genScript(info)

	cmd := exec.Command(getExpectPath(), defaultScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()

	os.Remove(defaultScriptPath)
}

func searchFromFzf(loginInfos []LoginInfo) int {

	buf := bytes.NewBuffer(nil)
	renderTable(loginInfos, buf)

	opts := fzf.DefaultOptions()
	opts.HeaderLines = headerLine
	opts.Reader = bytes.NewReader(buf.Bytes())
	idx := -1
	opts.Printer = func(s string) {
		idx, _ = strconv.Atoi(strings.TrimSpace(strings.Split(s, confSpliter)[0]))
	}
	opts.InlineInfo = true
	opts.SetHeight("30%")
	opts.SetLayoutReverse()
	fzf.PostProcessOptions(opts)

	fzf.Run(opts, "")

	return idx
}

func main() {

	homedir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("get home dir failed: %+v", err)
	}
	conffilePath := homedir + "/" + *conffile
	defaultScriptPath = homedir + "/" + defaultScript

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\nUsage : fastssh <host number>\n if there"+
			"s no host number, search will open\n", bar)
		flag.PrintDefaults()
	}

	flag.Parse()

	loginInfos := loadLoginInfoConf(conffilePath)

	/* host number mode */
	if len(os.Args) > 1 {
		idx, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalf("invalid arg, ")
		}

		if idx > len(loginInfos) {
			log.Fatalf("invalid arg, input index %d out of range, check your config (%s) or use search mode", idx, conffilePath)
		}
		runSSH(loginInfos[idx-1])
		return
	}

	/* search mode */
	idx := searchFromFzf(loginInfos)
	if idx == -1 {
		fmt.Println("no host selected")
		return
	}

	runSSH(loginInfos[idx-1])
}
