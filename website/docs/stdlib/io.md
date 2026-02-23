# IO & File System

## Console IO

```haira
import "io"

io.println("Hello, World!")    // Print with newline
io.print("Enter name: ")       // Print without newline
```

## File System

```haira
import "fs"

// Read file
content, err = fs.read_file("data.txt")
if err != nil {
    io.println("Error: ${err}")
}

// Write file
err = fs.write_file("output.txt", "Hello!")

// Check if file exists
exists = fs.exists("config.json")

// List directory
files, err = fs.read_dir("./data")
```

## Reading Uploaded Files

In workflows with `file` parameters:

```haira
@post("/upload")
workflow Upload(document: file) -> { content: string } {
    content, err = io.read_file(document)
    if err != nil {
        return { content: "Failed to read." }
    }
    return { content: content }
}
```
