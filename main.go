package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	_ "github.com/xeodou/go-sqlcipher"

	"github.com/AnimusPEXUS/utils/environ"

	"github.com/jinzhu/gorm"

	"golang.org/x/crypto/ssh/terminal"
)

type Data struct {
	gorm.Model
	Name string
	Text string
}

func displayHidden(txt string) (string, error) {

	e := environ.NewFromStrings(os.Environ())
	editor := e.Get("EDITOR", "mcedit")

	fn := "tmp.fl"

	err := ioutil.WriteFile(fn, []byte(txt), 0700)
	if err != nil {
		return "", err
	}

	c := exec.Command(editor, fn)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}

	err = os.Remove(fn)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

func askPass(prompt string) (string, error) {
	fmt.Printf("%s", prompt)
	defer fmt.Printf("\n")

	res, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func main() {

	defer fmt.Println("Bye!")

	fmt.Printf("")

	password := ""
	if p, err := askPass("Password?: "); err != nil {
		panic(err)
	} else {
		password = p
	}

	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	p := "PRAGMA key = '" + password + "';"
	err = db.Exec(p).Error
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Data{}).Error
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)

loo:
	for {

		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("\n")
				break
			}
			fmt.Println("error: " + err.Error())
			continue
		}

		command = strings.TrimRight(command, "\n")

		command_splitted := strings.Split(command, " ")

		switch command_splitted[0] {
		default:
			fmt.Println("unknown command")
		case "h":
			fallthrough
		case "help":
			fmt.Println(`
  h, help - help

  l       - list
  c name  - create
  e id    - edit
  v id    - view
  d id    - delete

  r       - change password
`)

		case "c":
			if len(command_splitted) != 2 {
				fmt.Println("name required")
				continue
			}

			err = db.Create(&Data{Name: command_splitted[1]}).Error
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}

		case "v":
			if len(command_splitted) != 2 {
				fmt.Println("id required")
				continue
			}

			var dat Data
			err = db.Where("id = ?", command_splitted[1]).First(&dat).Error
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}

			displayHidden(dat.Text)

		case "e":
			if len(command_splitted) != 2 {
				fmt.Println("id required")
				continue
			}

			var dat Data
			err = db.Where("id = ?", command_splitted[1]).First(&dat).Error
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}

			d, err := displayHidden(dat.Text)
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}

			err = db.Model(&dat).Update("Text", string(d)).Error
			if err != nil {
				fmt.Println("error: " + err.Error())
				continue
			}

		case "l":
			if len(command_splitted) != 1 {
				fmt.Println("no params")
				continue
			}

			lst2 := make([]*Data, 0)

			{
				lst := make([]*Data, 0)

				err := db.Find(&lst).Error
				if err != nil {
					fmt.Println("error: " + err.Error())
					continue
				}

				for _, i := range lst {
					lst2 = append(lst2, i)
				}
			}

			if len(lst2) > 1 {
				for i := 0; i != len(lst2)-1; i++ {
					for j := i + 1; j != len(lst2); j++ {
						if lst2[i].Name > lst2[j].Name {
							z := lst2[i]
							lst2[i] = lst2[j]
							lst2[j] = z
						}
					}
				}
			}

			for _, i := range lst2 {
				fmt.Printf("  id%03d '%s'\n", i.ID, i.Name)
			}

		case "d":
			if len(command_splitted) != 2 {
				fmt.Println("id required")
				continue
			}

			err = db.Where("id = ?", command_splitted[1]).Delete(&Data{}).Error
			if err != nil {
				fmt.Println(err)
				continue
			}

		case "r":
			if len(command_splitted) != 1 {
				fmt.Println("no params")
				continue
			}

			password1 := ""
			password2 := ""
			if p, err := askPass("RePassword?: "); err != nil {
				fmt.Println("error: " + err.Error())
				continue
			} else {
				password1 = p
			}

			if p, err := askPass("confirm: "); err != nil {
				fmt.Println("error: " + err.Error())
				continue
			} else {
				password2 = p
			}

			if password1 != password2 {
				fmt.Println("error: missmatch")
				continue
			}

			p := "PRAGMA rekey = '" + password1 + "';"
			err = db.Exec(p).Error
			if err != nil {
				fmt.Println(err)
				continue
			}

		case "exit":
			fallthrough
		case "quit":
			break loo
		}
	}

}