# Blog a Gogo

A static site generator and content tracker written in Go.

## Dependencies

- [YAML](https://github.com/go-yaml/yaml) 
- [blackfriday](https://github.com/russross/blackfriday)

## Usage

A small sample site is included. After building, run `./blog-a-gogo` to see the generated site running on port 8080 

Blog posts should use the extension `.post` and Markdown syntax. You can give Blog a Gogo some information about files in the files under the content directory by using a YAML block at the start of the file (aka, front matter). 

```yaml
---
title: My First Post
blurb: This is my very first post. 
date: 2016-01-010
---
```

To add an HTML page, create a file ending in `.tmpl` and place it in the content directory.

## Configuration

See `./blog-a-gogo -help` and `config.yml` for configuration options.  

## To-Do
- Unit testing
- Track the template directory
- Refactoring: Separate site generator/tracking from the server logic
