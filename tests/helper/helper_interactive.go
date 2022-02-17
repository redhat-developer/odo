package helper

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/kr/pty"
)

type Interactive struct {
	Command       []string
	ExpectFromtty []string
	SendOntty     []string
	err           error
}

//func RunInteractive(commonVar CommonVar, interVar Interactive) (string, error) {
func RunInteractive(commonVar CommonVar, interVar Interactive) (string, error) {

	// tmpdir, _ := ioutil.TempDir("", "")
	os.Chdir(commonVar.Context)
	// fmt.Println(tmpdir)
	// defer os.RemoveAll(tmpdir)

	ptm, pts, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}

	term := vt10x.New(vt10x.WithWriter(pts))

	c, err := expect.NewConsole(expect.WithStdin(ptm), expect.WithStdout(term), expect.WithCloser(pts, ptm))
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	//cmd := exec.Command("odo", "init")
	cmd := exec.Command(interVar.Command[0], interVar.Command[1:]...)
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(interVar.SendOntty); i++ {
		res, err := c.ExpectString(interVar.ExpectFromtty[i])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(buf, res)
		c.SendLine(interVar.SendOntty[i])
	}
	res, err := c.ExpectString(interVar.ExpectFromtty[len(interVar.ExpectFromtty)-1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(buf, res)
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	// Close the slave end of the pty, and read the remaining bytes from the master end.
	c.Tty().Close()

	fmt.Println(buf, err)
	return buf.String(), err
}
