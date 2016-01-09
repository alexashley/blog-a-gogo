# Blog a Gogo

Yet another static site generator in Go.

## Dependencies

- [YAML](https://github.com/go-yaml/yaml) 
- [blackfriday](https://github.com/russross/blackfriday)

## Usage

A small sample site is included. After building, run `./blog-a-gogo` to see the generated site running on port 8080 

Blog posts should use the extension `.post` and Markdown syntax.  

To add an HTML page, create a file ending in `.tmpl` and place it in the content directory.

## Configuration

See `./blog-a-gogo -help` and `config.yaml` for configuration options.  
