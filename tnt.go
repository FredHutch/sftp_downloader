package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kniren/gota/dataframe"
)

func deleteOlderFiles(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	patterns := []string{"Enrolamiento", "VisitSummary"}
	for _, pattern := range patterns {
		var max int
		var maxName string
		for _, file := range files {
			if strings.Index(file.Name(), pattern) > -1 {
				segs := strings.Split(file.Name(), "-")
				segs = strings.Split(segs[2], ".")
				num, err := strconv.Atoi(segs[0])
				if err != nil {
					return err
				}
				if num > max {
					max = num
					maxName = file.Name()
				}
			}
		}

		// now loop through again and delete all files that are not maxName
		for _, file := range files {
			if strings.Index(file.Name(), pattern) > -1 {
				if file.Name() != maxName {
					err = os.Remove(filepath.Join(path, file.Name()))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func moveFilesUp(rundir string) error {
	oldDir := filepath.Join(rundir, "REPORTE-CSV-CONTENEDOR", "REPORTE-TNTSTUDIES")
	files, err := ioutil.ReadDir(oldDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		err = os.Rename(filepath.Join(oldDir, f.Name()), filepath.Join(rundir, f.Name()))
		if err != nil {
			return err
		}
	}
	segs := strings.Split(oldDir, string(filepath.Separator))
	segs = segs[:len(segs)-1]
	upDir := strings.Join(segs, string(filepath.Separator))
	err = os.RemoveAll(upDir)
	if err != nil {
		return err
	}
	return nil
}

func renameFiles(rundir string) error {
	files, err := ioutil.ReadDir(rundir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.Index(f.Name(), "Enrolamiento") > -1 {
			err = os.Rename(filepath.Join(rundir, f.Name()), filepath.Join(rundir, "enr.TNT.csv"))
		} else if strings.Index(f.Name(), "VisitSummary") > -1 {
			err = os.Rename(filepath.Join(rundir, f.Name()), filepath.Join(rundir, "vs.TNT.csv"))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func convertTNTDates(rundir string) error {

	listing, err := ioutil.ReadDir(rundir)
	if err != nil {
		return err
	}
	for _, finfo := range listing {
		if finfo.IsDir() {
			continue
		}
		if !(strings.Index(finfo.Name(), "enr.TNT") > -1 || strings.Index(finfo.Name(), "vs.TNT") > -1) {
			continue
		}
		f, err := os.Open(filepath.Join(rundir, finfo.Name()))
		if err != nil {
			return err
		}
		newDf := dataframe.ReadCSV(f, dataframe.HasHeader(true), dataframe.DetectTypes(false))
		f.Close()
		newDf = newDf.Capply(convertDate)
		if strings.Index(finfo.Name(), "enr.TNT") > -1 {
			if stringInSlice("Pid", newDf.Names()) {
				newDf = newDf.Rename("PTID", "Pid")
			}
		} else if strings.Index(finfo.Name(), "vs.TNT") > -1 { // vs
			if stringInSlice("NroParticipante", newDf.Names()) {
				newDf = newDf.Rename("PTID", "NroParticipante")
			}
		}
		outfh, err := os.Create(filepath.Join(rundir, finfo.Name()))
		if err != nil {
			return err
		}
		err = newDf.WriteCSV(outfh, dataframe.WriteHeader(true))
		if err != nil {
			return err
		}
		err = outfh.Close()
		if err != nil {
			return err
		}

	}
	return nil
}
