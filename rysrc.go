// rysrc
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//getFilesList is to walk through the dir to get the file list
func getFilesList(dirName string, tpList []string) []string {
	fileList := make([]string, 0)

	err := filepath.Walk(dirName, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		//fmt.Println(path)

		for _, tplist := range tpList {

			if strings.HasSuffix(path, tplist) {
				fmt.Printf("%s has suffix %s\n", path, tplist)
				fileList = append(fileList, path)
			}

		}

		return nil

	})

	if err != nil {
		fmt.Printf("filepath.Walk error=%v\n", err)
	}

	//fmt.Printf("\nFileList:\n%v\n", fileList)
	return fileList
}

//GetFilesListChan is to walk through the dir to get the file list
//and packed into a chan
func GetFilesListChan(dirName string, chStr chan<- string) error {

	err := filepath.Walk(dirName, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		chStr <- path
		//fmt.Printf("%s\n", path)
		return nil

	})

	if err != nil {
		//fmt.Printf("filepath.Walk error=%v\n", err)
		log.Printf("filepath.Walk error=%v\n", err)
	}

	close(chStr)
	return err

}

//TarChanFiles will capture each files from the chan and tar into a package
func TarChanFiles(chStr <-chan string, tpList []string, tardir string) {

	for {
		if fileName, ok := <-chStr; ok {
			//fmt.Printf("%s\n", fileName)

			for _, tplist := range tpList {

				//fmt.Printf(" suffix is %s\n", tplist)

				if strings.HasSuffix(fileName, tplist) {
					fmt.Printf("%s has suffix %s\n", fileName, tplist)

					out, err := exec.Command("tar", "-rvf", tardir, fileName).Output()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("%s\n", out)

				}

			}
		} else {
			break
		}

	}

}

var usageInfo string = `This is a tool to tar the source code you need from some folder.

Usage:  %s [flags] dir_name

The following flags are recognized:
`

func usage() {
	fmt.Fprintf(os.Stderr, usageInfo, os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	var tarDir = flag.String("d", os.ExpandEnv("$HOME/rysrcTarPkg.tar"), "the dir to save the tar package")
	var fileType = flag.String("f", ".go", "Filetype such as:.c,.c++,.go ")
	var srcDir = flag.String("s", "../", "src dir to get the source files")
	var async = flag.Bool("asyn", false, "sync or async to output the tar files, recommand async for huge search")

	flag.Parse()
	if flag.NFlag() == 0 || flag.NArg() > 1 {
		usage()
	}

	fmt.Printf("dir=%s,type=%s,srcDir=%s, async=%v\n", *tarDir, *fileType, *srcDir, *async)
	//tmake([]string, 0)

	typeList := strings.Split(*fileType, ",")

	fmt.Printf("typelist is %q\n", typeList)
	if *async {
		runtime.GOMAXPROCS(2)
		chanStrCom := make(chan string)

		go GetFilesListChan(*srcDir, chanStrCom)

		TarChanFiles(chanStrCom, typeList, *tarDir)
	} else {
		fileList := getFilesList(*srcDir, typeList)

		for _, file := range fileList {
			//log.Printf("%s\n", file)
			out, err := exec.Command("tar", "-rvf", *tarDir, file).Output()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", out)

		}
	}

}
