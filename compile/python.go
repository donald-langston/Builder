package compile

import (
	"Builder/logger"
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

//Ruby creates zip from files passed in as arg
func Python() {

	//copies contents of .hidden to workspace
	hiddenDir := os.Getenv("BUILDER_HIDDEN_DIR")
	workspaceDir := os.Getenv("BUILDER_WORKSPACE_DIR")
	tempWorkspace := workspaceDir + "/temp/" 
	//make temp dir
	os.Mkdir(tempWorkspace, 0755)

	//add hidden dir contents to temp dir, install dependencies
	exec.Command("cp", "-a", hiddenDir+"/.", tempWorkspace).Run()

	//define dir path for command to run
	var fullPath string
	configPath := os.Getenv("BUILDER_DIR_PATH")
	//if user defined path in builder.yaml, full path is included in tempWorkspace, else add the local path 
	if (configPath != "") {
		fullPath = tempWorkspace
	} else {
		path, _ := os.Getwd()
		//combine local path to newly created tempWorkspace, gets rid of "." in path name
		fullPath = path + tempWorkspace[strings.Index(tempWorkspace, ".")+1:]
		// fmt.Println(path)
		// fmt.Println(fullPath)
	}
		
	//install dependencies/build, if yaml build type exists install accordingly
	buildTool := strings.ToLower(os.Getenv("BUILDER_BUILD_TOOL"))
	buildCmd := os.Getenv("BUILDER_BUILD_COMMAND")

	var cmd *exec.Cmd
	if buildCmd != "" {
		//user specified cmd
		cmd = exec.Command(buildCmd)
	} else if (buildTool == "pip") {
		fmt.Println(buildTool)
		cmd = exec.Command("pip3", "install", "-r", "requirements.txt", "-t", fullPath+"/requirements") 
    cmd.Dir = fullPath       // or whatever directory it's in
	} else {
		//default
		cmd = exec.Command("pip3", "install", "-r", "requirements.txt", "-t", fullPath+"/requirements") 
    cmd.Dir = fullPath      // or whatever directory it's in
	}
	//run cmd, check for err, log cmd
	logger.InfoLogger.Println(cmd)
	err := cmd.Run()
	if err != nil {
		logger.ErrorLogger.Println("Python project failed to compile.")
		fmt.Println(err)
		log.Fatal(err)
	}

	// Zip temp dir.
	outFile, err := os.Create(workspaceDir+"/temp.zip")
	if err != nil {
		 log.Fatal(err)
	}
	
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add files from temp dir to the archive.
	addPythonFiles(w, tempWorkspace, "")

	err = w.Close()
	if err != nil {
		logger.ErrorLogger.Println("Python project failed to compile.")
		 log.Fatal(err)
	}

	artifactPath := os.Getenv("BUILDER_OUTPUT_PATH")
	if (artifactPath != "") {
		exec.Command("cp", "-a", workspaceDir+"/temp.zip", artifactPath).Run()
	}
	logger.InfoLogger.Println("Python project compiled successfully.")
}

//recursively add files
func addPythonFiles(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
			fmt.Println(err)
	}

	for _, file := range files {
			if !file.IsDir() {
					dat, err := ioutil.ReadFile(basePath + file.Name())
					if err != nil {
							fmt.Println(err)
					}

					// Add some files to the archive.
					f, err := w.Create(baseInZip + file.Name())
					if err != nil {
							fmt.Println(err)
					}
					_, err = f.Write(dat)
					if err != nil {
							fmt.Println(err)
					}
			} else if file.IsDir() {

					// Recurse
					newBase := basePath + file.Name() + "/"
					addPythonFiles(w, newBase, baseInZip  + file.Name() + "/")
			}
	}
}