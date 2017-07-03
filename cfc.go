/**
 * cfc.go | Check for comments (JavaScript Files)
 *
 * Checks JavaScript files for comments above functions.
 * If a function doesn't have comment outline the function's
 * arguments and logic a comment is places above it.
 */

package main;

import (
    "fmt"
    "io/ioutil"
    "strings"
    "os"
    "regexp"
    "path/filepath"
    "sync"
    "bufio"
    "os/user"
);

func main() {

    // sync group used to make sure all async functions complete before finishing exec
    var wg sync.WaitGroup;

    // ask user for filepath to check .js files
    reader := bufio.NewReader(os.Stdin);
    fmt.Print("Enter directory to check: ");
    text, _ := reader.ReadString('\n');
    text = strings.Replace(text, "\n", "", -1);  // replace newline

    // check if path has a tilde in it for home directory.
    // if so replace with user's home directory
    if (strings.Contains(text, "~")) {
        usr, err := user.Current(); // get current user for
        if err != nil {
            fmt.Println(err)
        }
        text = strings.Replace(text, "~", usr.HomeDir, -1); // replace tilde with home directory
    }

    // walk file path and get all js files
    javascriptFiles := checkExt(".js", text);


    fmt.Println("Num of .js files: ", len(javascriptFiles));

    // for each .js file add and exec a go async function
    for index := 0; index < len(javascriptFiles); index++ {
        wg.Add(1); // add wait to wait group
        go addCommentsToFile(javascriptFiles[index], &wg); // add async func
    }

    wg.Wait(); // tell go to wait until all async functions are done

}

/**
 * Returns the string in between the start and end strings given.
 */
func GetStringInBetween(str string, start string, end string) (result string) {
    s := strings.Index(str, start)
    if s == -1 {
        return
    }
    s += len(start)
    e := strings.Index(str, end)
    return str[s:e]
}

/**
 * Checks filepath for the extension provided
 * - Ignores .git and node_modules
 */
func checkExt(ext string, basePath string) []string {
	pathS, err := filepath.Abs(basePath);
	if err != nil {
		panic(err);
	}
	var files []string;
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
        if f.IsDir() && (f.Name() == ".git" || f.Name() == "node_modules" ) {
            return filepath.SkipDir;
        }
		if !f.IsDir() {
			r, err := regexp.MatchString(ext, f.Name());
			if err == nil && r {
				files = append(files, path);
			}
		}
		return nil;
	})
	return files;
}

func addCommentsToFile(fileName string, wg *sync.WaitGroup) {

    // fmt.Println("Checking file: ", fileName);

    defer wg.Done()

    // return; // uncomment this for testing

    // read the file given
    b, err := ioutil.ReadFile(fileName); // just pass the file name
    if err != nil {
        fmt.Print(err);
        return;
    }

    var commentFlagged bool = false;
    var insideABlockComment bool = false;
    var hasWrittenToFile bool = false;

    str := string(b); // convert content to a 'string'
    line := 0;
    temp := strings.Split(str, "\n");

    for _, item := range temp {

        // if we're at a line with a function reset the comment flag
        if (commentFlagged && strings.Contains(item, "function")) {
            commentFlagged = false;

        // if we're at a function line and a comment wasn't flagged we need
        // to add one above the function with it's name
        } else if (strings.Contains(item, "function")) {
            functionName := GetStringInBetween(item, "function ", "()");
            commentBlock := "/**\n * " + functionName + "\n */\n";
            temp[line] = commentBlock + item;
            hasWrittenToFile = true;
        }

        // for checking whether we're in a block comment
        if (strings.Contains(item, "/*")) {
            insideABlockComment = true;
        }

        // and when we've left the comment block
        if (strings.Contains(item, "*/")) {
            insideABlockComment = false;
        }

        // turn comment flag off if the comment was inside a function or brackets from something else
        if (!insideABlockComment && strings.Contains(item, "}")) {
            commentFlagged = false;
        }

        // found a comment so flag it
        if (strings.Contains(item, "//") || strings.Contains(item, "/*")) {
            commentFlagged = true;
        }

        line++;
    }

    // write a new file if we've added a comment
    if (hasWrittenToFile) {
        err = ioutil.WriteFile(fileName, []byte(strings.Join(temp, "\n")), 0644)
        if err != nil {
            fmt.Print(err);
        }
        fmt.Println("New file for ", fileName, " created with comment blocks.");
    } else {
        // fmt.Println("No comments added. Well done!");
    }
}
