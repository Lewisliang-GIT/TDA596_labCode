package chord

import (
	"fmt"
	"os"
)

func (node *Node) storeChordFile(f File, backup bool) bool {
	// Store the file in the bucket
	// Return true if success, false if failed
	// Append the file to the bucket
	f.FileId.Mod(f.FileId, hashMod)
	// Check if the file is already in the bucket

	if backup {
		for k, _ := range node.Bucket {
			if k.Cmp(f.FileId) == 0 {
				fmt.Println("File already in the Backup")
				return false
			}
		}
		node.Bucket[f.FileId] = f.Name
		fmt.Println("Store Backup: ", node.Bucket)
	} else {
		for k, _ := range node.Bucket {
			if k.Cmp(f.FileId) == 0 {
				fmt.Println("File already in the Bucket")
				return false
			}
		}
		node.Bucket[f.FileId] = f.Name
		fmt.Println("Store Bucket: ", node.Bucket)
	}
	currentNodeFileDownloadPath := "../files/" + string(node.Address) + "/chord_storage/"
	filepath := currentNodeFileDownloadPath + f.Name
	// Create the file on file path and store content
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Println("Create file failed")
		return false
	}
	defer file.Close()
	_, err = file.Write(f.Content)
	if err != nil {
		fmt.Println("Write file failed")
		return false
	}
	// Store the file in the file download folder
	return true
}

func (node *Node) storeLocalFile(f File) bool {
	// Store the file in the bucket
	// Return true if success, false if failed
	// Append the file to the bucket
	f.FileId.Mod(f.FileId, hashMod)
	currentNodeFileDownloadPath := "../files/" + string(node.Address) + "/file_download/"
	filepath := currentNodeFileDownloadPath + f.Name
	// Create the file on file path and store content
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Println("Create file failed")
		return false
	}
	defer file.Close()
	_, err = file.Write(f.Content)
	if err != nil {
		fmt.Println("Write file failed")
		return false
	}
	// Store the file in the file download folder
	return true
}
