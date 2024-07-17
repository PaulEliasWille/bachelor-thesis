package main

import (
	"encoding/json"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/go-github/github"
	"github.com/yargevad/filepathx"
	"golang.org/x/exp/maps"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Supported managed t1 platforms:          AWS, Azure, GCP, IBM, Oracle, Digital Ocean, Alibaba, Tencent, Cloudflare, Fastly
// Supported managed t2 platforms:          Vercel, Netlify, Firebase
// Supported self-hosted non-k8s platforms: Fn Project, Nuclio
// Supported self-hosted k8s platforms:     OpenWhisk, Fission, Kubeless, OpenFaaS, Nuclio, Knative

// Tools that require looking at multiple indicators to identify FaaS applications:
// - https://www.terraform.io/ ; platforms: AWS, Azure, GCP, IBM, Oracle, Alibaba
// - https://kubernetes.io/de/ ; platforms: OpenWhisk, Fission, Kubeless, OpenFaaS, Nuclio, Knative
// - https://www.pulumi.com ; platforms: AWS, Azure, GCP, IBM, Oracle, Alibaba
// - https://aws.amazon.com/de/cloudformation ; platforms: AWS
// - https://learn.microsoft.com/de-de/azure/azure-resource-manager/management/overview ; platforms: Azure
// - https://cloud.google.com/deployment-manager/docs ; platforms: GCP
// - https://www.ibm.com/products/schematics ; platforms: IBM ; frameworks: terraform
// - https://www.alibabacloud.com/en/product/ros ; platforms: Alibaba
// - https://aws.amazon.com/de/cdk ; platforms: AWS
// - https://sst.dev ; platforms: AWS

// Specific config files that guarantee FaaS applications:
// - https://aws.amazon.com/serverless/sam ; platforms: AWS
// - https://serverless.com ; platforms: AWS
// - https://vercel.com ; platforms: Vercel
// - https://www.netlify.com ; platforms: Netlify
// - https://nitric.io ; platforms: AWS, Azure, GCP
// - https://arc.codes ; platforms: AWS
// - https://begin.com/ ; platform: AWS ; framework: architect
// - https://fnproject.io ; platforms: Fn Project

// # Relevant
// [✓] firebase
// [✓] fastly
// [✓] cloudflare
// [✓] tencent
// [✓] openfaas
// [✓] digital ocean functions

// # Irrelevant
// [x] begin (uses architect)
// [x] faas framework (not found)
// [x] nimbella (inactive)
// [x] tsuru (not serverless)
// [x] trigger mesh (relevant but to far out of scope)
// [x] iron functions (inactive)

var ExtensionToLanguage = map[string]string{
	".as":          "ActionScript",
	".ada":         "Ada",
	".adb":         "Ada",
	".ads":         "Ada",
	".alda":        "Alda",
	".Ant":         "Ant",
	".adoc":        "AsciiDoc",
	".asciidoc":    "AsciiDoc",
	".asm":         "Assembly",
	".S":           "Assembly",
	".s":           "Assembly",
	".dats":        "ATS",
	".sats":        "ATS",
	".hats":        "ATS",
	".ahk":         "AutoHotkey",
	".awk":         "Awk",
	".bat":         "Batch",
	".btm":         "Batch",
	".bicep":       "Bicep",
	".bb":          "BitBake",
	".be":          "Berry",
	".cairo":       "Cairo",
	".carbon":      "Carbon",
	".cbl":         "COBOL",
	".cmd":         "Batch",
	".bash":        "BASH",
	".sh":          "Bourne Shell",
	".c":           "C",
	".carp":        "Carp",
	".csh":         "C Shell",
	".ec":          "C",
	".erl":         "Erlang",
	".hrl":         "Erlang",
	".pgc":         "C",
	".capnp":       "Cap'n Proto",
	".chpl":        "Chapel",
	".circom":      "Circom",
	".cs":          "C#",
	".clj":         "Clojure",
	".coffee":      "CoffeeScript",
	".cfm":         "ColdFusion",
	".cfc":         "ColdFusion CFScript",
	".cmake":       "CMake",
	".cc":          "C++",
	".cpp":         "C++",
	".cxx":         "C++",
	".pcc":         "C++",
	".c++":         "C++",
	".cr":          "Crystal",
	".css":         "CSS",
	".cu":          "CUDA",
	".d":           "D",
	".dart":        "Dart",
	".dhall":       "Dhall",
	".dtrace":      "DTrace",
	".dts":         "Device Tree",
	".dtsi":        "Device Tree",
	".e":           "Eiffel",
	".elm":         "Elm",
	".el":          "LISP",
	".exp":         "Expect",
	".ex":          "Elixir",
	".exs":         "Elixir",
	".feature":     "Gherkin",
	".factor":      "Factor",
	".fish":        "Fish",
	".fr":          "Frege",
	".fst":         "F*",
	".F#":          "F#",   // deplicated F#/GLSL
	".GLSL":        "GLSL", // both use ext '.fs'
	".vs":          "GLSL",
	".shader":      "HLSL",
	".cg":          "HLSL",
	".cginc":       "HLSL",
	".hlsl":        "HLSL",
	".lean":        "Lean",
	".hlean":       "Lean",
	".lgt":         "Logtalk",
	".lisp":        "LISP",
	".lsp":         "LISP",
	".lua":         "Lua",
	".ls":          "LiveScript",
	".sc":          "LISP",
	".f":           "FORTRAN Legacy",
	".F":           "FORTRAN Legacy",
	".f77":         "FORTRAN Legacy",
	".for":         "FORTRAN Legacy",
	".ftn":         "FORTRAN Legacy",
	".pfo":         "FORTRAN Legacy",
	".f90":         "FORTRAN Modern",
	".F90":         "FORTRAN Modern",
	".f95":         "FORTRAN Modern",
	".f03":         "FORTRAN Modern",
	".f08":         "FORTRAN Modern",
	".gleam":       "Gleam",
	".g4":          "ANTLR",
	".go":          "Go",
	".go2":         "Go",
	".groovy":      "Groovy",
	".gradle":      "Groovy",
	".h":           "C Header",
	".hbs":         "Handlebars",
	".hs":          "Haskell",
	".hpp":         "C++ Header",
	".hh":          "C++ Header",
	".html":        "HTML",
	".ha":          "Hare",
	".hx":          "Haxe",
	".hxx":         "C++ Header",
	".idr":         "Idris",
	".imba":        "Imba",
	".il":          "SKILL",
	".ino":         "Arduino Sketch",
	".io":          "Io",
	".iss":         "Inno Setup",
	".ipynb":       "Jupyter Notebook",
	".jai":         "JAI",
	".java":        "Java",
	".jsp":         "JSP",
	".js":          "JavaScript",
	".mjs":         "JavaScript",
	".cjs":         "JavaScript",
	".ts":          "TypeScript",
	".jl":          "Julia",
	".janet":       "Janet",
	".json":        "JSON",
	".jsx":         "JSX",
	".kak":         "KakouneScript",
	".kk":          "Koka",
	".kt":          "Kotlin",
	".kts":         "Kotlin",
	".lds":         "LD Script",
	".less":        "LESS",
	".ly":          "Lilypond",
	".Objective-C": "Objective-C", // deplicated Obj-C/Matlab/Mercury
	".Matlab":      "MATLAB",      // both use ext '.m'
	".Mercury":     "Mercury",     // use ext '.m'
	".md":          "Markdown",
	".markdown":    "Markdown",
	".mo":          "Motoko",
	".Motoko":      "Motoko",
	".ne":          "Nearley",
	".nix":         "Nix",
	".nsi":         "NSIS",
	".nsh":         "NSIS",
	".nu":          "Nu",
	".ML":          "OCaml",
	".ml":          "OCaml",
	".mli":         "OCaml",
	".mll":         "OCaml",
	".mly":         "OCaml",
	".mm":          "Objective-C++",
	".maven":       "Maven",
	".makefile":    "Makefile",
	".meson":       "Meson",
	".mustache":    "Mustache",
	".m4":          "M4",
	".mojo":        "Mojo",
	".🔥":           "Mojo",
	".move":        "Move",
	".l":           "lex",
	".nim":         "Nim",
	".njk":         "Nunjucks",
	".odin":        "Odin",
	".ohm":         "Ohm",
	".php":         "PHP",
	".pas":         "Pascal",
	".PL":          "Perl",
	".pl":          "Perl",
	".pm":          "Perl",
	".plan9sh":     "Plan9 Shell",
	".pony":        "Pony",
	".ps1":         "PowerShell",
	".text":        "Plain Text",
	".txt":         "Plain Text",
	".polly":       "Polly",
	".proto":       "Protocol Buffers",
	".prql":        "PRQL",
	".py":          "Python",
	".pxd":         "Cython",
	".pyx":         "Cython",
	".q":           "Q",
	".qml":         "QML",
	".r":           "R",
	".R":           "R",
	".raml":        "RAML",
	".Rebol":       "Rebol",
	".red":         "Red",
	".rego":        "Rego",
	".Rmd":         "RMarkdown",
	".rake":        "Ruby",
	".rb":          "Ruby",
	".resx":        "XML resource", // ref: https://docs.microsoft.com/en-us/dotnet/framework/resources/creating-resource-files-for-desktop-apps#ResxFiles
	".ring":        "Ring",
	".rkt":         "Racket",
	".rhtml":       "Ruby HTML",
	".rs":          "Rust",
	".rst":         "ReStructuredText",
	".sass":        "Sass",
	".scala":       "Scala",
	".scss":        "Sass",
	".scm":         "Scheme",
	".sed":         "sed",
	".stan":        "Stan",
	".star":        "Starlark",
	".sml":         "Standard ML",
	".sol":         "Solidity",
	".sql":         "SQL",
	".svelte":      "Svelte",
	".swift":       "Swift",
	".t":           "Terra",
	".tex":         "TeX",
	".thy":         "Isabelle",
	".tla":         "TLA",
	".sty":         "TeX",
	".tcl":         "Tcl/Tk",
	".toml":        "TOML",
	".TypeScript":  "TypeScript",
	".tsx":         "TypeScript",
	".tf":          "HCL",
	".um":          "Umka",
	".mat":         "Unity-Prefab",
	".prefab":      "Unity-Prefab",
	".Coq":         "Coq",
	".vala":        "Vala",
	".Verilog":     "Verilog",
	".csproj":      "MSBuild script",
	".vbproj":      "MSBuild script",
	".vcproj":      "MSBuild script",
	".vb":          "Visual Basic",
	".vim":         "VimL",
	".vue":         "Vue",
	".vy":          "Vyper",
	".xml":         "XML",
	".XML":         "XML",
	".xsd":         "XSD",
	".xsl":         "XSLT",
	".xslt":        "XSLT",
	".wxs":         "WiX",
	".yaml":        "YAML",
	".yml":         "YAML",
	".y":           "Yacc",
	".yul":         "Yul",
	".zep":         "Zephir",
	".zig":         "Zig",
	".zsh":         "Zsh",
}

var ExtensionToCategory = map[string]string{
	".as":                      "Source Code",
	".ada":                     "Source Code",
	".adb":                     "Source Code",
	".ads":                     "Source Code",
	".alda":                    "Source Code",
	".Ant":                     "Source Code",
	".adoc":                    "Source Code",
	".asciidoc":                "Source Code",
	".asm":                     "Source Code",
	".S":                       "Source Code",
	".s":                       "Source Code",
	".dats":                    "Source Code",
	".sats":                    "Source Code",
	".hats":                    "Source Code",
	".ahk":                     "Source Code",
	".awk":                     "Source Code",
	".bat":                     "Source Code",
	".btm":                     "Source Code",
	".bicep":                   "Source Code",
	".bb":                      "Source Code",
	".be":                      "Source Code",
	".cairo":                   "Source Code",
	".carbon":                  "Source Code",
	".cbl":                     "Source Code",
	".cmd":                     "Source Code",
	".bash":                    "Source Code",
	".sh":                      "Source Code",
	".c":                       "Source Code",
	".carp":                    "Source Code",
	".csh":                     "Source Code",
	".ec":                      "Source Code",
	".erl":                     "Source Code",
	".hrl":                     "Source Code",
	".pgc":                     "Source Code",
	".capnp":                   "Source Code",
	".chpl":                    "Source Code",
	".circom":                  "Source Code",
	".cs":                      "Source Code",
	".clj":                     "Source Code",
	".coffee":                  "Source Code",
	".cfm":                     "Source Code",
	".cfc":                     "Source Code",
	".cmake":                   "Source Code",
	".cc":                      "Source Code",
	".cpp":                     "Source Code",
	".cxx":                     "Source Code",
	".pcc":                     "Source Code",
	".c++":                     "Source Code",
	".cr":                      "Source Code",
	".css":                     "Source Code",
	".cu":                      "Source Code",
	".d":                       "Source Code",
	".dart":                    "Source Code",
	".dhall":                   "Source Code",
	".dtrace":                  "Source Code",
	".dts":                     "Source Code",
	".dtsi":                    "Source Code",
	".e":                       "Source Code",
	".elm":                     "Source Code",
	".el":                      "Source Code",
	".exp":                     "Source Code",
	".ex":                      "Source Code",
	".exs":                     "Source Code",
	".feature":                 "Source Code",
	".factor":                  "Source Code",
	".fish":                    "Source Code",
	".fr":                      "Source Code",
	".fst":                     "Source Code",
	".F#":                      "Source Code",
	".GLSL":                    "Source Code",
	".vs":                      "Source Code",
	".shader":                  "Source Code",
	".cg":                      "Source Code",
	".cginc":                   "Source Code",
	".hlsl":                    "Source Code",
	".lean":                    "Source Code",
	".hlean":                   "Source Code",
	".lgt":                     "Source Code",
	".lisp":                    "Source Code",
	".lsp":                     "Source Code",
	".lua":                     "Source Code",
	".ls":                      "Source Code",
	".sc":                      "Source Code",
	".f":                       "Source Code",
	".F":                       "Source Code",
	".f77":                     "Source Code",
	".for":                     "Source Code",
	".ftn":                     "Source Code",
	".pfo":                     "Source Code",
	".f90":                     "Source Code",
	".F90":                     "Source Code",
	".f95":                     "Source Code",
	".f03":                     "Source Code",
	".f08":                     "Source Code",
	".gleam":                   "Source Code",
	".g4":                      "Source Code",
	".go":                      "Source Code",
	".go2":                     "Source Code",
	".groovy":                  "Source Code",
	".gradle":                  "Source Code",
	".h":                       "Source Code",
	".hbs":                     "Source Code",
	".hs":                      "Source Code",
	".hpp":                     "Source Code",
	".hh":                      "Source Code",
	".html":                    "Source Code",
	".ha":                      "Source Code",
	".hx":                      "Source Code",
	".hxx":                     "Source Code",
	".idr":                     "Source Code",
	".imba":                    "Source Code",
	".il":                      "Source Code",
	".ino":                     "Source Code",
	".io":                      "Source Code",
	".iss":                     "Source Code",
	".ipynb":                   "Source Code",
	".jai":                     "Source Code",
	".java":                    "Source Code",
	".jsp":                     "Source Code",
	".js":                      "Source Code",
	".mjs":                     "Source Code",
	".cjs":                     "Source Code",
	".ts":                      "Source Code",
	".jl":                      "Source Code",
	".janet":                   "Source Code",
	".json":                    "Data",
	".jsx":                     "Source Code",
	".kak":                     "Source Code",
	".kk":                      "Source Code",
	".kt":                      "Source Code",
	".kts":                     "Source Code",
	".lds":                     "Source Code",
	".less":                    "Source Code",
	".ly":                      "Source Code",
	".Objective-C":             "Source Code",
	".Matlab":                  "Source Code",
	".Mercury":                 "Source Code",
	".md":                      "Documentation",
	".markdown":                "Documentation",
	".mo":                      "Source Code",
	".Motoko":                  "Source Code",
	".ne":                      "Source Code",
	".nix":                     "Source Code",
	".nsi":                     "Source Code",
	".nsh":                     "Source Code",
	".nu":                      "Source Code",
	".ML":                      "Source Code",
	".ml":                      "Source Code",
	".mli":                     "Source Code",
	".mll":                     "Source Code",
	".mly":                     "Source Code",
	".mm":                      "Source Code",
	".maven":                   "Source Code",
	".makefile":                "Source Code",
	".meson":                   "Source Code",
	".mustache":                "Source Code",
	".m4":                      "Source Code",
	".mojo":                    "Source Code",
	".🔥":                       "Source Code",
	".move":                    "Source Code",
	".l":                       "Source Code",
	".nim":                     "Source Code",
	".njk":                     "Source Code",
	".odin":                    "Source Code",
	".ohm":                     "Source Code",
	".php":                     "Source Code",
	".pas":                     "Source Code",
	".PL":                      "Source Code",
	".pl":                      "Source Code",
	".pm":                      "Source Code",
	".plan9sh":                 "Source Code",
	".pony":                    "Source Code",
	".ps1":                     "Source Code",
	".text":                    "Source Code",
	".txt":                     "Data",
	".polly":                   "Source Code",
	".proto":                   "Source Code",
	".prql":                    "Source Code",
	".py":                      "Source Code",
	".pxd":                     "Source Code",
	".pyx":                     "Source Code",
	".q":                       "Source Code",
	".qml":                     "Source Code",
	".r":                       "Source Code",
	".R":                       "Source Code",
	".raml":                    "Source Code",
	".Rebol":                   "Source Code",
	".red":                     "Source Code",
	".rego":                    "Source Code",
	".Rmd":                     "Source Code",
	".rake":                    "Source Code",
	".rb":                      "Source Code",
	".resx":                    "Source Code",
	".ring":                    "Source Code",
	".rkt":                     "Source Code",
	".rhtml":                   "Source Code",
	".rs":                      "Source Code",
	".rst":                     "Source Code",
	".sass":                    "Source Code",
	".scala":                   "Source Code",
	".scss":                    "Source Code",
	".scm":                     "Source Code",
	".sed":                     "Source Code",
	".stan":                    "Source Code",
	".star":                    "Source Code",
	".sml":                     "Source Code",
	".sol":                     "Source Code",
	".sql":                     "Source Code",
	".svelte":                  "Source Code",
	".swift":                   "Source Code",
	".t":                       "Source Code",
	".tex":                     "Source Code",
	".thy":                     "Source Code",
	".tla":                     "Source Code",
	".sty":                     "Source Code",
	".tcl":                     "Source Code",
	".toml":                    "Data",
	".TypeScript":              "Source Code",
	".tsx":                     "Source Code",
	".tf":                      "Source Code",
	".um":                      "Source Code",
	".mat":                     "Source Code",
	".prefab":                  "Source Code",
	".Coq":                     "Source Code",
	".vala":                    "Source Code",
	".Verilog":                 "Source Code",
	".csproj":                  "Source Code",
	".vbproj":                  "Source Code",
	".vcproj":                  "Source Code",
	".vb":                      "Source Code",
	".vim":                     "Source Code",
	".vue":                     "Source Code",
	".vy":                      "Source Code",
	".xml":                     "Data",
	".XML":                     "Data",
	".xsd":                     "Data",
	".xsl":                     "Data",
	".xslt":                    "Data",
	".wxs":                     "Source Code",
	".yaml":                    "Data",
	".yml":                     "Data",
	".y":                       "Source Code",
	".yul":                     "Source Code",
	".zep":                     "Source Code",
	".zig":                     "Source Code",
	".zsh":                     "Source Code",
	".webp":                    "Asset",
	".png":                     "Asset",
	".snap":                    "Other",
	".pdf":                     "Asset",
	".gif":                     "Asset",
	".jpg":                     "Asset",
	".lock":                    "Other",
	".jpeg":                    "Asset",
	"":                         "Other",
	".ttf":                     "Asset",
	".zip":                     "Data",
	".xcf":                     "Other",
	".wasm":                    "Other",
	".mp4":                     "Asset",
	".svg":                     "Asset",
	".a":                       "Other",
	".JPG":                     "Asset",
	".sketch":                  "Other",
	".gem":                     "Other",
	".bpe":                     "Other",
	".eot":                     "Other",
	".webm":                    "Asset",
	".6":                       "Other",
	".otf":                     "Asset",
	".woff":                    "Asset",
	".1":                       "Other",
	".woff2":                   "Asset",
	".so":                      "Other",
	".log":                     "Other",
	".mov":                     "Asset",
	".mp3":                     "Asset",
	".gitignore":               "Other",
	".m4v":                     "Asset",
	".ejs":                     "Other",
	".graphql":                 "Source Code",
	".scssc":                   "Source Code",
	".PNG":                     "Asset",
	".ico":                     "Asset",
	".psd":                     "Other",
	".tmpl":                    "Other",
	".gql":                     "Source Code",
	".csv":                     "Data",
	".dita":                    "Other",
	".dll":                     "Other",
	".sum":                     "Other",
	".liquid":                  "Other",
	".orig":                    "Other",
	".mdx":                     "Other",
	".tpl":                     "Other",
	".excalidraw":              "Other",
	".vtl":                     "Other",
	".3":                       "Other",
	".DS_Store":                "Other",
	".babelrc":                 "Other",
	".dockerignore":            "Other",
	".eslintignore":            "Other",
	".prettierignore":          "Other",
	".prettierrc":              "Other",
	".nvmrc":                   "Other",
	".gitkeep":                 "Other",
	".state":                   "Other",
	".editorconfig":            "Other",
	".map":                     "Other",
	".npmignore":               "Other",
	".helmignore":              "Other",
	".webmanifest":             "Other",
	".env":                     "Other",
	".eslintrc":                "Other",
	".firebaserc":              "Other",
	".gitattributes":           "Other",
	".funcignore":              "Other",
	".npmrc":                   "Other",
	".template":                "Other",
	".mod":                     "Other",
	".hcl":                     "Other",
	".sample":                  "Other",
	".conf":                    "Other",
	".Dockerfile":              "Other",
	".properties":              "Other",
	".styl":                    "Other",
	".rules":                   "Other",
	".schema":                  "Other",
	".gcloudignore":            "Other",
	".arc":                     "Other",
	".plist":                   "Other",
	".patch":                   "Other",
	".browserslistrc":          "Other",
	".puml":                    "Other",
	".j2":                      "Other",
	".config":                  "Other",
	".arc-config":              "Other",
	".snyk":                    "Other",
	".whitesource":             "Other",
	".cfg":                     "Other",
	".ini":                     "Other",
	".jinja":                   "Other",
	".tfvars":                  "Other",
	".la":                      "Other",
	".pc":                      "Other",
	".tool-versions":           "Other",
	".enc":                     "Other",
	".vtt":                     "Other",
	".pem":                     "Other",
	".sln":                     "Other",
	".terraform-version":       "Other",
	".jshintrc":                "Other",
	".git-blame-ignore-revs":   "Other",
	".development":             "Other",
	".all-contributorsrc":      "Other",
	".ditamap":                 "Other",
	".flake8":                  "Other",
	".keep":                    "Other",
	".mdlrc":                   "Other",
	".adr-dir":                 "Other",
	".ics":                     "Other",
	".iml":                     "Other",
	".rule":                    "Other",
	".tfbackend":               "Other",
	".swcrc":                   "Other",
	".docx":                    "Other",
	".xlsx":                    "Other",
	".pug":                     "Other",
	".local":                   "Other",
	".storyboard":              "Other",
	".xcconfig":                "Other",
	".xcsettings":              "Other",
	".xcworkspacedata":         "Other",
	".commented":               "Other",
	".njsproj":                 "Other",
	".project":                 "Other",
	".pydevproject":            "Other",
	".flowconfig":              "Other",
	".gitmodules":              "Other",
	".15":                      "Other",
	".62":                      "Other",
	".erb":                     "Other",
	".envrc":                   "Other",
	".htc":                     "Other",
	".jshintignore":            "Other",
	".mermaid":                 "Other",
	".terraformignore":         "Other",
	".pptx":                    "Other",
	".pub_key":                 "Other",
	".c8rc":                    "Other",
	".sequelizerc":             "Other",
	".vercelignore":            "Other",
	".code-workspace":          "Other",
	".production":              "Other",
	".staging":                 "Other",
	".node":                    "Other",
	".server":                  "Other",
	".codespellrc":             "Other",
	".fql":                     "Other",
	".pws":                     "Other",
	".work":                    "Other",
	".cabal":                   "Other",
	".exemple":                 "Other",
	".nojekyll":                "Other",
	".filters":                 "Other",
	".vcxproj":                 "Other",
	".ai":                      "Other",
	".stage":                   "Other",
	".releaserc":               "Other",
	".api":                     "Other",
	".display":                 "Other",
	".listener":                "Other",
	".eleventyignore":          "Other",
	".stylelintrc":             "Other",
	".node-version":            "Other",
	".cache":                   "Other",
	".sqlite":                  "Other",
	".suo":                     "Other",
	".jsbeautifyrc":            "Other",
	".ignore":                  "Other",
	".tftpl":                   "Other",
	".prisma":                  "Other",
	".postcssrc":               "Other",
	".export_metadata":         "Other",
	".overall_export_metadata": "Other",
	".2":                       "Other",
	".7":                       "Other",
	".pdb":                     "Other",
	".dat":                     "Other",
	".entitlements":            "Other",
	".metadata":                "Other",
	".pbxproj":                 "Other",
	".xcscheme":                "Other",
	".editorConfig":            "Other",
	".env-sample":              "Other",
	".dio":                     "Other",
	".commitlintrc":            "Other",
}

type SingleLineCommentSymbol string

type MultiLineCommentSymbol struct {
	Start string
	End   string
}

type CommentSymbols struct {
	SingleLine []SingleLineCommentSymbol
	MultiLine  []MultiLineCommentSymbol
}

var LanguageToCommentSymbols = map[string]CommentSymbols{
	"ActionScript":        CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Ada":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Alda":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Ant":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"ANTLR":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"AsciiDoc":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Assembly":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//", ";", "#", "@", "|", "!"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"ATS":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}, {Start: "(*", End: "*)"}}},
	"AutoHotkey":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Awk":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Arduino Sketch":      CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Batch":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"REM", "rem"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Berry":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "#-", End: "-#"}}},
	"BASH":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Bicep":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"BitBake":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"C":                   CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"C Header":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"C Shell":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Cairo":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Carbon":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Cap'n Proto":         CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Carp":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"C#":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Chapel":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Circom":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Clojure":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#", "#_"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"COBOL":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"*", "/"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"CoffeeScript":        CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "###", End: "###"}}},
	"Coq":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"(*"}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"ColdFusion":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "<!---", End: "--->"}}},
	"ColdFusion CFScript": CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"CMake":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"C++":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"C++ Header":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Crystal":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"CSS":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Cython":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "\"\"\"", End: "\"\"\""}}},
	"CUDA":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"D":                   CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Dart":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//", "///"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Dhall":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "{-", End: "-}"}}},
	"DTrace":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Device Tree":         CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Eiffel":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Elm":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "{-", End: "-}"}}},
	"Elixir":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Erlang":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"%"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Expect":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Fish":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Frege":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "{-", End: "-}"}}},
	"F*":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{"(*", "//"}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"F#":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{"(*"}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"Lean":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "/-", End: "-/"}}},
	"Logtalk":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"%"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Lua":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "--[[", End: "]]"}}},
	"Lilypond":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"%"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"LISP":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{";;"}, MultiLine: []MultiLineCommentSymbol{{Start: "#|", End: "|#"}}},
	"LiveScript":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Factor":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"! "}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"FORTRAN Legacy":      CommentSymbols{SingleLine: []SingleLineCommentSymbol{"c", "C", "!", "*"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"FORTRAN Modern":      CommentSymbols{SingleLine: []SingleLineCommentSymbol{"!"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Gherkin":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Gleam":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"GLSL":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Go":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Groovy":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Handlebars":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}, {Start: "{{!", End: "}}"}}},
	"Haskell":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "{-", End: "-}"}}},
	"Haxe":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Hare":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"HLSL":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"HTML":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//", "<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Idris":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "{-", End: "-}"}}},
	"Imba":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "###", End: "###"}}},
	"Io":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//", "#"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"SKILL":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"JAI":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Janet":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Java":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"JSP":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"JavaScript":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Julia":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "#:=", End: ":=#"}}},
	"Jupyter Notebook":    CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"JSON":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"JSX":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"KakouneScript":       CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Koka":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Kotlin":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"LD Script":           CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"LESS":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Objective-C":         CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Markdown":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Motoko":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Nearley":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Nix":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"NSIS":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#", ";"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Nu":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{";", "#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"OCaml":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"Objective-C++":       CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Makefile":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"MATLAB":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"%"}, MultiLine: []MultiLineCommentSymbol{{Start: "%{", End: "}%"}}},
	"Mercury":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"%"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Maven":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Meson":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Mojo":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Move":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Mustache":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "{{!", End: "}}"}}},
	"M4":                  CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Nim":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "#[", End: "]#"}}},
	"Nunjucks":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "{#", End: "#}"}, {Start: "<!--", End: "-->"}}},
	"lex":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Odin":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Ohm":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"PHP":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#", "//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Pascal":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "{", End: ")"}}},
	"Perl":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: ":=", End: ":=cut"}}},
	"Plain Text":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Plan9 Shell":         CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Pony":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"PowerShell":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "<#", End: "#>"}}},
	"Polly":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Protocol Buffers":    CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"PRQL":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Python":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "\"\"\"", End: "\"\"\""}}},
	"Q":                   CommentSymbols{SingleLine: []SingleLineCommentSymbol{"/ "}, MultiLine: []MultiLineCommentSymbol{{Start: "\\", End: "/"}, {Start: "/", End: "\\"}}},
	"QML":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"R":                   CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Rebol":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Red":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Rego":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"RMarkdown":           CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"RAML":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Racket":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "#|", End: "|#"}}},
	"ReStructuredText":    CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Ring":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#", "//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Ruby":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: ":=Start: begin", End: "End: :=end"}}},
	"Ruby HTML":           CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Rust":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//", "///", "//!"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Scala":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Sass":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Scheme":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "#|", End: "|#"}}},
	"sed":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Stan":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Starlark":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Solidity":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Bourne Shell":        CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Standard ML":         CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"SQL":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Svelte":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}, {Start: "<!--", End: "-->"}}},
	"Swift":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Terra":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"--"}, MultiLine: []MultiLineCommentSymbol{{Start: "--[[", End: "]]"}}},
	"TeX":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"%"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Inno Setup":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{";"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Isabelle":            CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"TLA":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"\\*"}, MultiLine: []MultiLineCommentSymbol{{Start: "(*", End: "*)"}}},
	"Tcl/Tk":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"TOML":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"TypeScript":          CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"HCL":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#", "//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Umka":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Unity-Prefab":        CommentSymbols{SingleLine: []SingleLineCommentSymbol{}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"MSBuild script":      CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Vala":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Verilog":             CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"VimL":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{`"`}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Visual Basic":        CommentSymbols{SingleLine: []SingleLineCommentSymbol{"'"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Vue":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"Vyper":               CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "\"\"\"", End: "\"\"\""}}},
	"WiX":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"XML":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"XML resource":        CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"XSLT":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"XSD":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"<!--"}, MultiLine: []MultiLineCommentSymbol{{Start: "<!--", End: "-->"}}},
	"YAML":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Yacc":                CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Yul":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Zephir":              CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//"}, MultiLine: []MultiLineCommentSymbol{{Start: "/*", End: "*/"}}},
	"Zig":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"//", "///"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
	"Zsh":                 CommentSymbols{SingleLine: []SingleLineCommentSymbol{"#"}, MultiLine: []MultiLineCommentSymbol{{Start: "", End: ""}}},
}

type FaaSPlatform string

const (
	FaaSPlatformUnknown FaaSPlatform = "unknown"

	FaaSPlatformCloudflare   FaaSPlatform = "cloudflare"
	FaaSPlatformFastly       FaaSPlatform = "fastly"
	FaaSPlatformAWS          FaaSPlatform = "aws"
	FaaSPlatformGCP          FaaSPlatform = "gcp"
	FaaSPlatformFirebase     FaaSPlatform = "firebase"
	FaaSPlatformAzure        FaaSPlatform = "azure"
	FaaSPlatformIBM          FaaSPlatform = "ibm"
	FaaSPlatformOracle       FaaSPlatform = "oracle"
	FaaSPlatformAlibaba      FaaSPlatform = "alibaba"
	FaaSPlatformTencent      FaaSPlatform = "tencent"
	FaaSPlatformDigitalOcean FaaSPlatform = "digitalocean"

	FaaSPlatformVercel  FaaSPlatform = "vercel"
	FaaSPlatformNetlify FaaSPlatform = "netlify"

	FaaSPlatformFnProject FaaSPlatform = "fnproject"
	FaaSPlatformNuclio    FaaSPlatform = "nuclio"
	FaaSPlatformOpenWhisk FaaSPlatform = "openwhisk"
	FaaSPlatformFission   FaaSPlatform = "fission"
	FaaSPlatformKubeless  FaaSPlatform = "kubeless"
	FaaSPlatformKnative   FaaSPlatform = "knative"
	FaaSPlatformOpenFaaS  FaaSPlatform = "openfaas"
)

type FaaSFramework string

const (
	FaaSFrameworkUnknown FaaSFramework = "unknown"

	FaaSFrameworkVercel  FaaSFramework = "vercel"
	FaaSFrameworkNetlify FaaSFramework = "netlify"

	FaaSFrameworkFastly                   FaaSFramework = "fastly"
	FaaSFrameworkWrangler                 FaaSFramework = "wrangler"
	FaaSFrameworkServerless               FaaSFramework = "serverless"
	FaaSFrameworkFirebase                 FaaSFramework = "firebase"
	FaaSFrameworkNitric                   FaaSFramework = "nitric"
	FaaSFrameworkArchitect                FaaSFramework = "architect"
	FaaSFrameworkAWSCDKAndSST             FaaSFramework = "aws_cdk_and_sst"
	FaaSFrameworkGCPFunctions             FaaSFramework = "gcp_functions"
	FaaSFrameworkAzureFunctions           FaaSFramework = "azure_functions"
	FaaSFrameworkAzureDurableFunctions    FaaSFramework = "azure_durable_functions"
	FaaSFrameworkServerlessCloudFramework FaaSFramework = "serverless_cloud_framework"
	FaaSFrameworkDigitalOcean             FaaSFramework = "digitalocean"
	FaaSFrameworkAlexaSkillsKit           FaaSFramework = "alexa_skills_kit"
	FaaSFrameworkHono                     FaaSFramework = "hono"

	FaaSFrameworkTerraform                           FaaSFramework = "terraform"
	FaaSFrameworkPulumi                              FaaSFramework = "pulumi"
	FaaSFrameworkAWSCloudFormationAndSAM             FaaSFramework = "aws_cloudformation_and_sam"
	FaaSFrameworkAzureResourceManager                FaaSFramework = "azure_resource_manager"
	FaaSFrameworkGCPCloudDeploymentManager           FaaSFramework = "gcp_cloud_deployment_manager"
	FaaSFrameworkAlibabaResourceOrchestrationService FaaSFramework = "alibaba_resource_orchestration_service"

	FaaSFrameworkFnProject FaaSFramework = "fnproject"
	FaaSFrameworkNuclio    FaaSFramework = "nuclio"
	FaaSFrameworkOpenWhisk FaaSFramework = "openwhisk"
	FaaSFrameworkFission   FaaSFramework = "fission"
	FaaSFrameworkKubeless  FaaSFramework = "kubeless"
	FaaSFrameworkKnative   FaaSFramework = "knative"
	FaaSFrameworkOpenFaaS  FaaSFramework = "openfaas"
)

type FaaSInvocationType string

const (
	FaaSInvocationTypeUnknown   FaaSInvocationType = "unknown"
	FaaSInvocationTypeHTTP      FaaSInvocationType = "http"
	FaaSInvocationTypeWebsocket FaaSInvocationType = "websocket"
	FaaSInvocationTypeGraphQL   FaaSInvocationType = "graphql"
	FaaSInvocationTypeSchedule  FaaSInvocationType = "schedule"
	FaaSInvocationTypeTopic     FaaSInvocationType = "topic"
	FaaSInvocationTypeQueue     FaaSInvocationType = "queue"
	FaaSInvocationTypeOther     FaaSInvocationType = "other"
)

type FaaSLocation string

const (
	FaaSLocationUnknown FaaSLocation = "unknown"
	FaaSLocationEdge    FaaSLocation = "edge"
	FaaSLocationRegion  FaaSLocation = "region"
)

type RepositoryPackageData struct {
	RootPath string

	Name        string
	Description string

	Dependencies            []string
	DevDependencies         []string
	FaaSRuntimeDependencies []string

	PublishedToNPM bool
	NumPackages    int

	NumFilesByExtension    map[string]int
	LinesOfTextByExtension map[string]int

	NumFaaSHandlers            int
	NumFaaSRuntimeDependencies int
}

type RepositoryFaaSFunctionData struct {
	Name string

	Platform       FaaSPlatform
	Framework      FaaSFramework
	InvocationType FaaSInvocationType
	Location       FaaSLocation

	TimeoutSeconds int

	SourceFilePath string
	SourceFileLine int
}

type RepositoryComplexityDataFile struct {
	Extension     string
	Name          string
	Path          string
	Language      string
	Category      string
	LOC           int
	NumCharacters int
}

type RepositoryComplexityData struct {
	Files []RepositoryComplexityDataFile
}

type RepositoryData struct {
	RepositoryId RepositoryId
	Url          string

	Name        string
	Description string
	License     string
	Topics      []string
	Stars       int
	Watchers    int
	Forks       int

	Size     int
	Archived bool
	Forked   bool

	CreatedAt time.Time
	PushedAt  time.Time

	NumContributors int
	NumIssues       int
	NumOpenIssues   int
	NumClosedIssues int

	NumCommits         int
	FirstCommitAt      time.Time
	FirstHumanCommitAt time.Time
	LastCommitAt       time.Time
	LastHumanCommitAt  time.Time
	ActiveDays         int
	ActiveHumanDays    int

	Complexity RepositoryComplexityData

	UsedPlatforms  map[FaaSPlatform]bool
	UsedFrameworks map[FaaSFramework]bool

	Functions    []RepositoryFaaSFunctionData
	NumFunctions int

	Packages                   []RepositoryPackageData
	NumPackages                int
	NumPublishedToNPM          int
	Dependencies               []string
	DevDependencies            []string
	FaaSRuntimeDependencies    []string
	NumFaaSRuntimeDependencies int
	NumFaaSHandlers            int
}

type TextFile struct {
	Path      string
	Extension string
	Content   string
	NumLines  int
}

func LoadTextFiles(includePattern string, excludeDirectories []string) ([]TextFile, error) {
	matches, err := filepathx.Glob(includePattern)
	if err != nil {
		return nil, err
	}

	result := make([]TextFile, 0, len(matches))

	for _, match := range matches {
		fileInfo, err := os.Stat(match)
		if err != nil {
			fmt.Printf("error finding file: %v\n", err)
			continue
		}

		if fileInfo.IsDir() {
			continue
		}

		if checkForKeywords(match, excludeDirectories) > 0 {
			continue
		}

		contentBytes, err := os.ReadFile(match)
		if err != nil {
			fmt.Printf("error reading file: %v\n", err)
			continue
		}

		content := string(contentBytes)
		numLines := strings.Count(content, "\n")
		extension := path.Ext(match)

		result = append(result, TextFile{
			Path:      match,
			Extension: extension,
			Content:   content,
			NumLines:  numLines,
		})
	}
	return result, nil
}

func FilterTextFiles(
	textFiles []TextFile,
	includePatterns ...string,
) ([]TextFile, error) {
	textFilePaths := make([]string, len(textFiles))
	for _, textFile := range textFiles {
		textFilePaths = append(textFilePaths, textFile.Path)
	}

	filteredTextFilePaths := make([]string, 0)
	for _, textFilePath := range textFilePaths {
		for _, includePattern := range includePatterns {
			matches, err := doublestar.Match(includePattern, textFilePath)
			if err != nil {
				continue
			}

			if matches {
				filteredTextFilePaths = append(filteredTextFilePaths, textFilePath)
				break
			}
		}
	}

	filteredTextFiles := make([]TextFile, 0)
	for _, textFile := range textFiles {
		if slices.Contains(filteredTextFilePaths, textFile.Path) {
			filteredTextFiles = append(filteredTextFiles, textFile)
		}
	}
	return filteredTextFiles, nil
}

func usedPlatformsToString(platforms map[FaaSPlatform]bool) string {
	results := make([]string, 0)
	for platform, isUsed := range platforms {
		if isUsed {
			results = append(results, string(platform))
		}
	}
	return strings.Join(results, ";")
}

func usedFrameworksToString(frameworks map[FaaSFramework]bool) string {
	results := make([]string, 0)
	for framework, isUsed := range frameworks {
		if isUsed {
			results = append(results, string(framework))
		}
	}
	return strings.Join(results, ";")
}

func usedLocationsToString(locations map[FaaSLocation]bool) string {
	results := make([]string, 0)
	for location, isUsed := range locations {
		if isUsed {
			results = append(results, string(location))
		}
	}
	return strings.Join(results, ";")
}

func extractComplexity(files []TextFile) RepositoryComplexityData {
	repoCompDataFiles := make([]RepositoryComplexityDataFile, 0, len(files))

	for _, file := range files {
		language := file.Extension
		if value, ok := ExtensionToLanguage[file.Extension]; ok {
			language = value
		}

		category := file.Extension
		if value, ok := ExtensionToCategory[file.Extension]; ok {
			category = value
		}

		//commentSymbols, ok := LanguageToCommentSymbols[language]
		//if !ok {
		//	commentSymbols = CommentSymbols{
		//		SingleLine: []SingleLineCommentSymbol{},
		//		MultiLine:  []MultiLineCommentSymbol{},
		//	}

		repoCompDataFile := RepositoryComplexityDataFile{
			Extension:     file.Extension,
			Name:          path.Base(file.Path),
			Path:          file.Path,
			Language:      language,
			Category:      category,
			LOC:           0,
			NumCharacters: 0,
		}

		if category == "Source Code" {
			lines := strings.Split(file.Content, "\n")

			//var insideMultilineComment bool = false

			for _, rawLine := range lines {
				line := strings.TrimSpace(rawLine)

				if len(line) == 0 {
					continue
				}

				repoCompDataFile.LOC += 1
				repoCompDataFile.NumCharacters += len(line)

				//isCode := false
				//
				//linePos := 0
				//for linePos < len(line) {
				//	if !insideMultilineComment {
				//		singleLineCommentStart := -1
				//		for _, symbol := range commentSymbols.SingleLine {
				//			idx := strings.Index(line[linePos:], string(symbol)) + linePos
				//			if idx != -1 && (singleLineCommentStart == -1 || singleLineCommentStart > idx) {
				//				singleLineCommentStart = idx
				//			}
				//		}
				//
				//		multiLineCommentStart := -1
				//		for _, symbols := range commentSymbols.MultiLine {
				//			idx := strings.Index(line[linePos:], symbols.Start) + linePos
				//			if idx != -1 && (multiLineCommentStart == -1 || multiLineCommentStart > idx) {
				//				multiLineCommentStart = idx
				//			}
				//		}
				//
				//		switch {
				//		case singleLineCommentStart == -1 && multiLineCommentStart == -1:
				//			isCode = true
				//			break
				//		case multiLineCommentStart != -1 && singleLineCommentStart == -1:
				//			if multiLineCommentStart > 0 {
				//				isCode = true
				//			}
				//
				//		case singleLineCommentStart != -1 && singleLineCommentStart < multiLineCommentStart:
				//			// pass
				//		}
				//
				//		if isSingleLineComment {
				//			continue
				//		}
				//	}
				//}
				//
				//if isCode {
				//	locByLanguage[language] += 1
				//}
			}
		}

		if repoCompDataFile.LOC == 1 && repoCompDataFile.NumCharacters > 120 {
			continue
		}

		if repoCompDataFile.LOC > 2000 {
			continue
		}

		if repoCompDataFile.NumCharacters > 2000*120 {
			continue
		}

		repoCompDataFiles = append(repoCompDataFiles, repoCompDataFile)
	}

	return RepositoryComplexityData{
		Files: repoCompDataFiles,
	}
}

func checkForKeywords(src string, keywords []string) int {
	result := 0
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(src), strings.ToLower(keyword)) {
			result += 1
		}
	}
	return result
}

func durationInDays(duration time.Duration) int {
	hours := duration.Hours()
	return int(math.Round(hours / 24))
}

func filterHumanCommits(commits []github.RepositoryCommit) []github.RepositoryCommit {
	filteredCommits := make([]github.RepositoryCommit, 0)
	for _, commit := range commits {
		// TODO: not all commits have an author specified, so to ensure that we dont filter commits
		//       that are from humans, we accept that some bot commits make it through
		if commit.GetAuthor().GetType() == "Bot" {
			continue
		}
		filteredCommits = append(filteredCommits, commit)
	}
	return filteredCommits
}

func getFirstAndLastCommitDate(commits []github.RepositoryCommit) (time.Time, time.Time) {
	first := time.Time{}
	last := time.Time{}

	for _, commit := range commits {
		current := commit.GetCommit().GetAuthor().GetDate()

		if first.IsZero() || first.After(current) {
			first = current
		}

		if last.IsZero() || last.Before(current) {
			last = current
		}
	}

	return first, last
}

func checkIfPublishedToNpm(httpClient *http.Client, repositoryInfo *github.Repository, packageJson interface{}) bool {
	packageName, err := JsonResolveString(packageJson, []string{"name"})
	if err != nil {
		return false
	}

	npmManifest, err := RetryWithResult(func() (interface{}, error) {
		return DownloadJson(httpClient, fmt.Sprintf("https://registry.npmjs.org/%s", packageName))
	}, DefaultErrorHandler)
	if err != nil {
		return false
	}

	npmRepositoryType, err := JsonResolveString(npmManifest, []string{"repository", "type"})
	if err != nil {
		return false
	}

	if npmRepositoryType != "git" {
		return false
	}

	npmRepositoryUrl, err := JsonResolveString(npmManifest, []string{"repository", "url"})
	if err != nil {
		return false
	}

	if !strings.HasSuffix(npmRepositoryUrl, repositoryInfo.GetCloneURL()) {
		return false
	}

	return true
}

func scanPackageJson(data *RepositoryPackageData, packageJsonFile TextFile) {
	data.Name = ""
	data.Description = ""
	data.Dependencies = make([]string, 0)
	data.DevDependencies = make([]string, 0)

	packageJson, err := LoadJsonFromBytes([]byte(packageJsonFile.Content))
	if err != nil {
		fmt.Printf("error loading packageJson: %v\n", err)
		return
	}

	if name, err := JsonResolveString(packageJson, []string{"name"}); err == nil {
		data.Name = name
	} else {
		data.Name = ""
	}

	if description, err := JsonResolveString(packageJson, []string{"description"}); err == nil {
		data.Description = description
	} else {
		data.Description = ""
	}

	if dependenciesMap, err := JsonResolveMap(packageJson, []string{"dependencies"}); err == nil {
		data.Dependencies = maps.Keys(dependenciesMap)
	} else {
		data.Dependencies = []string{}
	}

	if devDependenciesMap, err := JsonResolveMap(packageJson, []string{"devDependencies"}); err == nil {
		data.DevDependencies = maps.Keys(devDependenciesMap)
	} else {
		data.DevDependencies = []string{}
	}

	// TODO: uncomment this
	// data.PublishedToNPM = checkIfPublishedToNpm(httpClient, &repositoryInfo, packageJson)
}

func scanFaaSRuntimeDependencies(data *RepositoryPackageData) {
	faasRuntimeDependencies := []string{
		"@google-cloud/functions-framework",
		"firebase-functions",
		"@azure/functions",
		"@ibm-functions/composer",
		"@fnproject/fdk",
		"serverless",
		"@architect/functions",
	}

	data.FaaSRuntimeDependencies = IntersectionSlice(faasRuntimeDependencies, data.Dependencies)
	data.NumFaaSRuntimeDependencies = len(data.FaaSRuntimeDependencies)
}

func scanFaaSHandlers(data *RepositoryPackageData, files []TextFile) {
	for _, file := range files {
		if !(file.Extension == ".js" || file.Extension == ".mjs") {
			continue
		}
		data.NumFaaSHandlers += strings.Count(file.Content, "exports.handler")
		data.NumFaaSHandlers += strings.Count(file.Content, "export const handler")
	}
}

func scanPackageJsons(data *RepositoryPackageData, files []TextFile) {
	for _, file := range files {
		if strings.HasSuffix(file.Path, "package.json") {
			data.NumPackages += 1
		}
	}
}

func scanVercel(data *RepositoryData, files []TextFile) {
	vercelConfigFiles, err := FilterTextFiles(files, "**/vercel.json")
	if err != nil {
		return
	}

	for _, vercelConfigFile := range vercelConfigFiles {
		vercelConfig, err := LoadJsonFromBytes([]byte(vercelConfigFile.Content))
		if err != nil {
			continue
		}

		data.UsedFrameworks[FaaSFrameworkVercel] = true
		data.UsedPlatforms[FaaSPlatformVercel] = true

		functionBaseDirectory := path.Dir(vercelConfigFile.Path)

		functionDirectories := []string{
			path.Join(functionBaseDirectory, "api/**/*.js"),
			path.Join(functionBaseDirectory, "api/**/*.mjs"),
			path.Join(functionBaseDirectory, "pages/api/**/*.js"),
			path.Join(functionBaseDirectory, "pages/api/**/*.mjs"),
			path.Join(functionBaseDirectory, "src/pages/api/**/*.js"),
			path.Join(functionBaseDirectory, "src/pages/api/**/*.mjs"),
		}

		functions, err := JsonResolveMap(vercelConfig, []string{"functions"})
		if err == nil {
			for functionDirectory, _ := range functions {
				functionDirectories = append(functionDirectories, path.Join(functionBaseDirectory, functionDirectory))
			}
		}

		jsFiles, err := FilterTextFiles(
			files,
			functionDirectories...,
		)

		if err != nil {
			continue
		}

		for _, jsFile := range jsFiles {
			hasHandler1 := strings.Contains(jsFile.Content, "export default") && strings.Contains(jsFile.Content, "handler")
			hasHandler2 := strings.Contains(jsFile.Content, "module.exports") && (strings.Contains(jsFile.Content, "req") ||
				strings.Contains(jsFile.Content, "res"))
			hasHandler3 := strings.Contains(jsFile.Content, "export function GET(") ||
				strings.Contains(jsFile.Content, "export function POST(") ||
				strings.Contains(jsFile.Content, "export function PUT(") ||
				strings.Contains(jsFile.Content, "export function DELETE(") ||
				strings.Contains(jsFile.Content, "export function HEAD(") ||
				strings.Contains(jsFile.Content, "export function OPTIONS(")

			if !(hasHandler1 || hasHandler2 || hasHandler3) {
				continue
			}

			function := RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformVercel,
				Framework:      FaaSFrameworkVercel,
				InvocationType: FaaSInvocationTypeHTTP,
				Location:       FaaSLocationUnknown,
				TimeoutSeconds: -1,
				SourceFilePath: jsFile.Path,
				SourceFileLine: -1,
			}

			isEdge1 := strings.Contains(jsFile.Content, "export const config") &&
				strings.Contains(jsFile.Content, "runtime") &&
				(strings.Contains(jsFile.Content, "'edge'") || strings.Contains(jsFile.Content, "\"edge\""))

			isEdge2 := strings.Contains(jsFile.Content, "export const runtime") &&
				(strings.Contains(jsFile.Content, "'edge'") || strings.Contains(jsFile.Content, "\"edge\""))

			if isEdge1 || isEdge2 {
				function.Location = FaaSLocationEdge
			} else {
				function.Location = FaaSLocationRegion
			}

			// TODO: peak into vercel.json to get possible cron schedules and additional runtime configuration

			data.Functions = append(data.Functions, function)
		}
	}
}

func scanNetlify(data *RepositoryData, files []TextFile) {
	netlifyConfigFiles, err := FilterTextFiles(
		files,
		"**/netlify.toml",
	)
	if err != nil {
		return
	}

	for _, netlifyConfigFile := range netlifyConfigFiles {
		netlifyConfig, err := LoadJsonFromTomlBytes([]byte(netlifyConfigFile.Content))
		if err != nil {
			continue
		}

		data.UsedFrameworks[FaaSFrameworkNetlify] = true
		data.UsedPlatforms[FaaSPlatformNetlify] = true

		buildBase, err := JsonResolveString(netlifyConfig, []string{"build", "base"})
		if err != nil {
			buildBase = ""
		}

		functionsBase := ""
		if v1, err := JsonResolveString(netlifyConfig, []string{"functions", "directory"}); err == nil {
			functionsBase = v1
		} else if v2, err := JsonResolveString(netlifyConfig, []string{"build", "functions"}); err == nil {
			functionsBase = v2
		} else {
			functionsBase = "netlify/functions"
		}

		functionsPath := path.Join(
			path.Dir(netlifyConfigFile.Path),
			buildBase,
			functionsBase,
		)

		functionsJsFiles, err := FilterTextFiles(
			files,
			path.Join(functionsPath, "**/*.js"),
			path.Join(functionsPath, "**/*.mjs"),
		)
		if err == nil {
			for _, jsFile := range functionsJsFiles {
				if !(strings.Contains(jsFile.Content, "exports.handler") ||
					strings.Contains(jsFile.Content, "export const handler") ||
					strings.Contains(jsFile.Content, "export const function handler") ||
					strings.Contains(jsFile.Content, "export const async function handler")) {
					continue
				}

				data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
					Name:           "",
					Platform:       FaaSPlatformNetlify,
					Framework:      FaaSFrameworkNetlify,
					InvocationType: FaaSInvocationTypeHTTP,
					Location:       FaaSLocationRegion,
					TimeoutSeconds: -1,
					SourceFilePath: jsFile.Path,
					SourceFileLine: -1,
				})
			}

			// TODO: Functions could also be defined elsewhere and only be put in the functions directory
			//       during the build process. We would need to somehow find the real location, maybe by
			//       scanning for build tools that target the functions directory location.
		}

		edgeFunctionsBase := ""
		if v, err := JsonResolveString(netlifyConfig, []string{"build", "edge_functions"}); err == nil {
			edgeFunctionsBase = v
		} else {
			edgeFunctionsBase = "netlify/edge_functions"
		}

		edgeFunctionsPath := path.Join(
			path.Dir(netlifyConfigFile.Path),
			buildBase,
			edgeFunctionsBase,
		)

		edgeFunctionsJsFiles, err := FilterTextFiles(
			files,
			path.Join(edgeFunctionsPath, "**/*.js"),
			path.Join(edgeFunctionsPath, "**/*.mjs"),
		)
		if err == nil {
			for _, jsFile := range edgeFunctionsJsFiles {
				if !(strings.Contains(jsFile.Content, "exports.handler") ||
					strings.Contains(jsFile.Content, "export const handler") ||
					strings.Contains(jsFile.Content, "export const function handler") ||
					strings.Contains(jsFile.Content, "export const async function handler")) {
					continue
				}

				data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
					Name:           "",
					Platform:       FaaSPlatformNetlify,
					Framework:      FaaSFrameworkNetlify,
					InvocationType: FaaSInvocationTypeHTTP,
					Location:       FaaSLocationEdge,
					TimeoutSeconds: -1,
					SourceFilePath: jsFile.Path,
					SourceFileLine: -1,
				})
			}

			// TODO: Edge functions could also be defined elsewhere and only be put in the functions directory
			//       during the build process. We would need to somehow find the real location, maybe by
			//       scanning for build tools that target the functions directory location.
		}
	}
}

func scanServerless(data *RepositoryData, files []TextFile) {
	serverlessConfigs, err := FilterTextFiles(
		files,
		"**/serverless.yml", "**/serverless.yaml",
	)
	if err != nil || len(serverlessConfigs) == 0 {
		return
	}

	data.UsedPlatforms[FaaSPlatformAWS] = true
	data.UsedFrameworks[FaaSFrameworkServerless] = true

	for _, serverlessConfig := range serverlessConfigs {
		// TODO: resolve file references (`${file(...)`}

		serverlessConfigJsons := LoadJsonsFromYamlBytes([]byte(serverlessConfig.Content))
		if len(serverlessConfigJsons) != 1 {
			continue
		}

		serverlessConfigJson := serverlessConfigJsons[0]

		// TODO: look at the resources to identify serverless resources
		// TODO: handle common (popular) serverless plugins

		serverlessFunctions, err := JsonResolveMap(serverlessConfigJson, []string{"functions"})
		if err != nil {
			continue
		}

		for _, serverlessFunction := range serverlessFunctions {
			function := RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformAWS,
				Framework:      FaaSFrameworkServerless,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationUnknown, // TODO: implement
				TimeoutSeconds: -1,
				SourceFilePath: serverlessConfig.Path,
				SourceFileLine: -1,
			}

			events, err := JsonResolveArray(serverlessFunction, []string{"events"})
			if err == nil {
				for _, event := range events {
					eventMap, err := JsonResolveMap(event, []string{})
					if err != nil {
						continue
					}

					eventNames := maps.Keys(eventMap)
					for _, eventName := range eventNames {
						switch eventName {
						case "httpApi", "http":
							function.InvocationType = FaaSInvocationTypeHTTP
							function.Location = FaaSLocationRegion
						case "websocket":
							function.InvocationType = FaaSInvocationTypeWebsocket
							function.Location = FaaSLocationRegion
						case "schedule":
							function.InvocationType = FaaSInvocationTypeSchedule
							function.Location = FaaSLocationRegion
						case "sns", "stream", "msk", "kafka":
							function.InvocationType = FaaSInvocationTypeTopic
							function.Location = FaaSLocationRegion
						case "sqs", "activemq", "rabbitmq":
							function.InvocationType = FaaSInvocationTypeQueue
							function.Location = FaaSLocationRegion
						case "s3", "alexa", "iot",
							"cloudwatchEvent", "cognitoUserPool",
							"alb", "eventBridge":
							function.InvocationType = FaaSInvocationTypeOther
							function.Location = FaaSLocationRegion
						case "cloudFront":
							function.InvocationType = FaaSInvocationTypeOther
							function.Location = FaaSLocationEdge
						default:
							fmt.Printf("unkown serverless event: %s\n", eventName)
							function.InvocationType = FaaSInvocationTypeUnknown
							function.Location = FaaSLocationUnknown
						}
					}
				}
			}

			data.Functions = append(data.Functions, function)
		}
	}
}

func scanNitric(data *RepositoryData, files []TextFile) {
	nitricConfigs, err := FilterTextFiles(
		files,
		"**/nitric.yaml",
		"**/nitric.yml",
	)
	if err != nil || len(nitricConfigs) == 0 {
		return
	}

	data.UsedPlatforms[FaaSPlatformAWS] = true
	data.UsedPlatforms[FaaSPlatformGCP] = true
	data.UsedPlatforms[FaaSPlatformAzure] = true
	data.UsedFrameworks[FaaSFrameworkNitric] = true

	for _, nitricConfig := range nitricConfigs {
		nitricConfigJsons := LoadJsonsFromYamlBytes([]byte(nitricConfig.Content))
		if len(nitricConfigJsons) != 1 {
			continue
		}
		nitricConfigJson := nitricConfigJsons[0]

		nitricServices, err := JsonResolveArray(nitricConfigJson, []string{"services"})
		if err != nil {
			continue
		}

		for _, nitricService := range nitricServices {
			nitricServiceRelMatchPattern, err := JsonResolveString(nitricService, []string{"match"})
			if err != nil {
				continue
			}

			nitricServiceFunctions, err := FilterTextFiles(
				files,
				path.Join(path.Dir(nitricConfig.Path), nitricServiceRelMatchPattern),
			)
			if err != nil {
				continue
			}

			for _, _ = range nitricServiceFunctions {
				// TODO: inspect function to figure out invocation type
				data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
					Name:           "",
					Platform:       FaaSPlatformAWS,
					Framework:      FaaSFrameworkServerless,
					InvocationType: FaaSInvocationTypeUnknown,
					Location:       FaaSLocationRegion,
					TimeoutSeconds: -1,
					SourceFilePath: nitricConfig.Path,
					SourceFileLine: -1,
				})
			}
		}
	}
}

func scanArchitect(data *RepositoryData, files []TextFile) {
	architectConfigs, err := FilterTextFiles(
		files,
		"**/.arc", "**/app.arc",
		"**/arc.yaml", "**/arc.yml",
		"**/arc.json",
	)
	if err != nil {
		return
	}

	if len(architectConfigs) == 0 {
		return
	}

	data.UsedPlatforms[FaaSPlatformAWS] = true
	data.UsedFrameworks[FaaSFrameworkArchitect] = true

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformAWS,
		Framework:      FaaSFrameworkArchitect,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationRegion,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}
	for _, architectConfig := range architectConfigs {
		if architectConfig.Extension == ".arc" {
			section := ""
			for _, line := range strings.Split(architectConfig.Content, "\n") {
				line = strings.TrimSpace(line)
				function := defaultFunction
				function.SourceFilePath = architectConfig.Path
				switch {
				case len(line) == 0:
					continue
				case strings.HasPrefix(line, "#"):
					continue
				case strings.HasPrefix(line, "@"):
					section = line[1:]
					continue
				case section == "ws":
					function.InvocationType = FaaSInvocationTypeWebsocket
				case section == "http":
					function.InvocationType = FaaSInvocationTypeHTTP
				case section == "queues":
					function.InvocationType = FaaSInvocationTypeQueue
				case section == "events":
					function.InvocationType = FaaSInvocationTypeTopic
				case section == "scheduled":
					function.InvocationType = FaaSInvocationTypeSchedule
				case section == "tables-streams":
					function.InvocationType = FaaSInvocationTypeOther
				default:
					continue
				}
				data.Functions = append(data.Functions, function)
			}
		} else {
			var architectConfigJson interface{}
			switch architectConfig.Extension {
			case ".yml", ".yaml":
				architectConfigJsons := LoadJsonsFromYamlBytes([]byte(architectConfig.Content))
				if len(architectConfigJsons) != 1 {
					continue
				}
				architectConfigJson = architectConfigJsons[0]
			case ".json":
				architectConfigJson, err = LoadJsonFromBytes([]byte(architectConfig.Content))
				if err != nil {
					continue
				}
			default:
				continue
			}

			architectWsFunctions, err := JsonResolveArray(architectConfigJson, []string{"ws"})
			if err == nil {
				for _, _ = range architectWsFunctions {
					function := defaultFunction
					function.SourceFilePath = architectConfig.Path
					function.InvocationType = FaaSInvocationTypeWebsocket
					data.Functions = append(data.Functions, function)
				}
			}

			architectHttpFunctions, err := JsonResolveArray(architectConfigJson, []string{"http"})
			if err == nil {
				for _, _ = range architectHttpFunctions {
					function := defaultFunction
					function.SourceFilePath = architectConfig.Path
					function.InvocationType = FaaSInvocationTypeHTTP
					data.Functions = append(data.Functions, function)
				}
			}

			architectQueueFunctions, err := JsonResolveArray(architectConfigJson, []string{"queues"})
			if err == nil {
				for _, _ = range architectQueueFunctions {
					function := defaultFunction
					function.SourceFilePath = architectConfig.Path
					function.InvocationType = FaaSInvocationTypeQueue
					data.Functions = append(data.Functions, function)
				}
			}

			architectPubSubFunctions, err := JsonResolveArray(architectConfigJson, []string{"events"})
			if err == nil {
				for _, _ = range architectPubSubFunctions {
					function := defaultFunction
					function.SourceFilePath = architectConfig.Path
					function.InvocationType = FaaSInvocationTypeTopic
					data.Functions = append(data.Functions, function)
				}
			}

			architectScheduledFunctions, err := JsonResolveArray(architectConfigJson, []string{"scheduled"})
			if err == nil {
				for _, _ = range architectScheduledFunctions {
					function := defaultFunction
					function.SourceFilePath = architectConfig.Path
					function.InvocationType = FaaSInvocationTypeSchedule
					data.Functions = append(data.Functions, function)
				}
			}

			architectOtherFunctions, err := JsonResolveArray(architectConfigJson, []string{"tables-streams"})
			if err == nil {
				for _, _ = range architectOtherFunctions {
					function := defaultFunction
					function.SourceFilePath = architectConfig.Path
					function.InvocationType = FaaSInvocationTypeOther
					data.Functions = append(data.Functions, function)
				}
			}
		}
	}
}

func scanAWSCDKAndSST(data *RepositoryData, files []TextFile) {
	if !(slices.Contains(data.Dependencies, "sst") ||
		slices.Contains(data.DevDependencies, "sst") ||
		slices.Contains(data.Dependencies, "aws-cdk-lib") ||
		slices.Contains(data.DevDependencies, "aws-cdk-lib")) {
		return
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	if len(jsFiles) == 0 {
		return
	}

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformAWS,
		Framework:      FaaSFrameworkAWSCDKAndSST,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationUnknown,       // TODO: implement
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	for _, jsFile := range jsFiles {
		if !(strings.Contains(jsFile.Content, "sst/constructs") ||
			strings.Contains(jsFile.Content, "@serverless-stack/resources") ||
			strings.Contains(jsFile.Content, "aws-cdk-lib")) {
			continue
		}

		numLambdas := strings.Count(jsFile.Content, "new Function(") +
			strings.Count(jsFile.Content, ".Function(")
		numNodejsLambdas := strings.Count(jsFile.Content, "new NodejsFunction(") +
			strings.Count(jsFile.Content, ".NodejsFunction(")
		for index := 0; index < numLambdas+numNodejsLambdas; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeUnknown
			data.Functions = append(data.Functions, function)
		}

		numApis := strings.Count(jsFile.Content, "new Api(") +
			strings.Count(jsFile.Content, ".Api(")
		for index := 0; index < numApis; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeHTTP
			function.Location = FaaSLocationRegion
			data.Functions = append(data.Functions, function)
		}

		numWebsocketApis := strings.Count(jsFile.Content, "new WebsocketApi(") +
			strings.Count(jsFile.Content, ".WebsocketApi(")
		for index := 0; index < numWebsocketApis; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeWebsocket
			function.Location = FaaSLocationRegion
			data.Functions = append(data.Functions, function)
		}

		numSchedules := strings.Count(jsFile.Content, "new Cron(") +
			strings.Count(jsFile.Content, ".Cron(")
		for index := 0; index < numSchedules; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeSchedule
			function.Location = FaaSLocationRegion
			data.Functions = append(data.Functions, function)
		}

		numJobs := strings.Count(jsFile.Content, "new Job(") +
			strings.Count(jsFile.Content, ".Job(")
		numQueues := strings.Count(jsFile.Content, "new Queue(") +
			strings.Count(jsFile.Content, ".Queue(")
		for index := 0; index < numJobs+numQueues; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeQueue
			function.Location = FaaSLocationRegion
			data.Functions = append(data.Functions, function)
		}

		numTopics := strings.Count(jsFile.Content, "new Topic(") +
			strings.Count(jsFile.Content, ".Topic(")
		numEventBus := strings.Count(jsFile.Content, "new EventBus(") +
			strings.Count(jsFile.Content, ".EventBus(")
		numKinesisStreams := strings.Count(jsFile.Content, "new KinesisStream(") +
			strings.Count(jsFile.Content, ".KinesisStream(")
		for index := 0; index < numTopics+numEventBus+numKinesisStreams; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeTopic
			function.Location = FaaSLocationRegion
			data.Functions = append(data.Functions, function)
		}

		numAppSyncApis := strings.Count(jsFile.Content, "new AppSyncApi(") +
			strings.Count(jsFile.Content, ".AppSyncApi(")
		for index := 0; index < numAppSyncApis; index++ {
			function := defaultFunction
			function.SourceFilePath = jsFile.Path
			function.InvocationType = FaaSInvocationTypeOther
			function.Location = FaaSLocationRegion
			data.Functions = append(data.Functions, function)
		}

		if numLambdas+numNodejsLambdas+numApis+numAppSyncApis+numWebsocketApis+
			numJobs+numSchedules+numTopics+numQueues+numEventBus+numKinesisStreams > 0 {
			data.UsedPlatforms[FaaSPlatformAWS] = true
			data.UsedFrameworks[FaaSFrameworkAWSCDKAndSST] = true
		}
	}
}

func scanTerraform(data *RepositoryData, files []TextFile) {
	terraformFiles, err := FilterTextFiles(
		files,
		"**/*.tf",
	)
	if err != nil {
		return
	}

	if len(terraformFiles) == 0 {
		return
	}

	platformChecks := map[string]FaaSPlatform{
		"resource \"aws_lambda_function\"":            FaaSPlatformAWS,
		"resource \"google_cloudfunctions_function\"": FaaSPlatformGCP,
		"resource \"azurerm_function_app\"":           FaaSPlatformAzure,
		"resource \"azurerm_linux_function_app\"":     FaaSPlatformAzure,
		"resource \"azurerm_windows_function_app\"":   FaaSPlatformAzure,
		"resource \"ibm_function_action\"":            FaaSPlatformIBM,
		"resource \"oci_functions_function\"":         FaaSPlatformOracle,
		"resource \"alicloud_fc_function\"":           FaaSPlatformAlibaba,
	}

	knownPlatformLocations := map[FaaSPlatform]FaaSLocation{
		FaaSPlatformGCP:     FaaSLocationRegion,
		FaaSPlatformAzure:   FaaSLocationRegion,
		FaaSPlatformIBM:     FaaSLocationRegion,
		FaaSPlatformOracle:  FaaSLocationRegion,
		FaaSPlatformAlibaba: FaaSLocationRegion,
	}

	for _, terraformFile := range terraformFiles {
		for check, platform := range platformChecks {
			count := strings.Count(terraformFile.Content, check)
			for i := 0; i < count; i++ {
				data.UsedFrameworks[FaaSFrameworkTerraform] = true
				data.UsedPlatforms[platform] = true

				location, ok := knownPlatformLocations[platform]
				if !ok {
					location = FaaSLocationUnknown
				}

				data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
					Name:           "",
					Platform:       platform,
					Framework:      FaaSFrameworkTerraform,
					InvocationType: FaaSInvocationTypeUnknown,
					Location:       location,
					TimeoutSeconds: -1,
					SourceFilePath: terraformFile.Path,
					SourceFileLine: -1,
				})
			}
		}
	}
}

func scanPulumi(data *RepositoryData, files []TextFile) {
	pulumiConfigFiles, err := FilterTextFiles(
		files,
		"**/Pulumi.yaml", "**/Pulumi.yml",
	)
	if err != nil {
		return
	}

	if len(pulumiConfigFiles) == 0 {
		return
	}

	allDependencies := append(data.Dependencies, data.DevDependencies...)

	if !(slices.Contains(allDependencies, "@pulumi/pulumi")) {
		return
	}

	data.UsedFrameworks[FaaSFrameworkPulumi] = true
	switch {
	case slices.Contains(allDependencies, "@pulumi/aws"),
		slices.Contains(allDependencies, "@pulumi/aws-native"),
		slices.Contains(allDependencies, "@pulumi/awsx"):
		data.UsedPlatforms[FaaSPlatformAWS] = true
	case slices.Contains(allDependencies, "@pulumi/azure-native"),
		slices.Contains(allDependencies, "@pulumi/azure"),
		slices.Contains(allDependencies, "@pulumi/azapi"):
		data.UsedPlatforms[FaaSPlatformAzure] = true
	case slices.Contains(allDependencies, "@pulumi/gcp"),
		slices.Contains(allDependencies, "@pulumi/google-native"):
		data.UsedPlatforms[FaaSPlatformGCP] = true
	case slices.Contains(allDependencies, "@pulumi/oci"):
		data.UsedPlatforms[FaaSPlatformOracle] = true
	case slices.Contains(allDependencies, "@pulumi/alicloud"):
		data.UsedPlatforms[FaaSPlatformAlibaba] = true
	default:
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	if len(jsFiles) == 0 {
		return
	}

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformUnknown, // set below
		Framework:      FaaSFrameworkPulumi,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationUnknown,       // set below
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	for _, jsFile := range jsFiles {
		switch {
		case strings.Contains(jsFile.Content, "@pulumi/aws"),
			strings.Contains(jsFile.Content, "@pulumi/aws-native"),
			strings.Contains(jsFile.Content, "@pulumi/awsx"):
			numFunctions := strings.Count(jsFile.Content, "lambda.Function(")
			for index := 0; index < numFunctions; index++ {
				function := defaultFunction
				function.Platform = FaaSPlatformAWS
				function.InvocationType = FaaSInvocationTypeUnknown
				function.Location = FaaSLocationUnknown
				function.SourceFilePath = jsFile.Path
				data.Functions = append(data.Functions, function)
			}
		case strings.Contains(jsFile.Content, "@pulumi/azure-native"),
			strings.Contains(jsFile.Content, "@pulumi/azure"),
			strings.Contains(jsFile.Content, "@pulumi/azapi"):
			numFunctions := strings.Count(jsFile.Content, "appservice.FunctionApp(") +
				strings.Count(jsFile.Content, "appservice.LinuxFunctionApp(") +
				strings.Count(jsFile.Content, "appservice.WindowsFunctionApp(")
			for index := 0; index < numFunctions; index++ {
				function := defaultFunction
				function.Platform = FaaSPlatformAzure
				function.InvocationType = FaaSInvocationTypeUnknown
				function.Location = FaaSLocationRegion
				function.SourceFilePath = jsFile.Path
				data.Functions = append(data.Functions, function)
			}
		case strings.Contains(jsFile.Content, "@pulumi/gcp"),
			strings.Contains(jsFile.Content, "@pulumi/google-native"):
			numFunctions := strings.Count(jsFile.Content, "cloudfunctions.Function(") +
				strings.Count(jsFile.Content, "cloudfunctionsv2.Function(")
			for index := 0; index < numFunctions; index++ {
				function := defaultFunction
				function.Platform = FaaSPlatformGCP
				function.InvocationType = FaaSInvocationTypeUnknown
				function.Location = FaaSLocationRegion
				function.SourceFilePath = jsFile.Path
				data.Functions = append(data.Functions, function)
			}
		case strings.Contains(jsFile.Content, "@pulumi/oci"):
			numFunctions := strings.Count(jsFile.Content, "functions.Function(")
			for index := 0; index < numFunctions; index++ {
				function := defaultFunction
				function.Platform = FaaSPlatformOracle
				function.InvocationType = FaaSInvocationTypeUnknown
				function.Location = FaaSLocationRegion
				function.SourceFilePath = jsFile.Path
				data.Functions = append(data.Functions, function)
			}
		case strings.Contains(jsFile.Content, "@pulumi/alicloud"):
			numFunctions := strings.Count(jsFile.Content, "fc.Function(")
			for index := 0; index < numFunctions; index++ {
				function := defaultFunction
				function.Platform = FaaSPlatformAlibaba
				function.InvocationType = FaaSInvocationTypeUnknown
				function.Location = FaaSLocationRegion
				function.SourceFilePath = jsFile.Path
				data.Functions = append(data.Functions, function)
			}
		default:
		}
	}
}

func scanAWSCloudFormationAndSAM(data *RepositoryData, files []TextFile) {
	configs, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
		"**/*.json",
	)
	if err != nil {
		return
	}

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformAWS,
		Framework:      FaaSFrameworkAWSCloudFormationAndSAM,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationUnknown,       // set below
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	numEdgeLambdaFunctions := 0
	prelimFunctions := make([]RepositoryFaaSFunctionData, 0)

	for _, config := range configs {
		var configJson interface{}
		switch config.Extension {
		case ".yaml", ".yml":
			configJsons := LoadJsonsFromYamlBytes([]byte(config.Content))
			if len(configJsons) != 1 {
				continue
			}
			configJson = configJsons[0]
		case ".json":
			configJson, err = LoadJsonFromBytes([]byte(config.Content))
			if err != nil {
				continue
			}
		default:
			continue
		}

		resources, err := JsonResolveMap(configJson, []string{"Resources"})
		if err != nil {
			continue
		}

		for _, resource := range resources {
			resourceType, err := JsonResolveString(resource, []string{"Type"})
			if err != nil {
				continue
			}

			data.UsedPlatforms[FaaSPlatformAWS] = true
			data.UsedFrameworks[FaaSFrameworkAWSCloudFormationAndSAM] = true

			switch resourceType {
			case "AWS::Lambda::Function":
				function := defaultFunction
				function.InvocationType = FaaSInvocationTypeUnknown
				function.Location = FaaSLocationRegion // corrected below
				function.SourceFilePath = config.Path
				prelimFunctions = append(prelimFunctions, function)
			case "AWS::Serverless::Function":
				function := defaultFunction
				function.SourceFilePath = config.Path
				function.Location = FaaSLocationRegion

				events, err := JsonResolveMap(resource, []string{"Properties", "Events"})
				if err == nil {
					for _, event := range events {
						eventType, err := JsonResolveString(event, []string{"Type"})
						if err != nil {
							continue
						}

						switch eventType {
						case "Api", "HttpApi":
							function.InvocationType = FaaSInvocationTypeHTTP
						case "Schedule", "ScheduleV2":
							function.InvocationType = FaaSInvocationTypeSchedule
						case "Kinesis", "SNS", "MQMSK", "SelfManagedKafka":
							function.InvocationType = FaaSInvocationTypeTopic
						case "SQS":
							function.InvocationType = FaaSInvocationTypeQueue
						case "S3", "DynamoDB", "AlexaSkill", "Cognito",
							"CloudWatchLogs", "CloudWatchEvent",
							"CognitoDocumentDB", "EventBridgeRule", "IoTRule":
							function.InvocationType = FaaSInvocationTypeOther
						default:
							fmt.Printf("eventType: %s\n", eventType)
							function.InvocationType = FaaSInvocationTypeUnknown
						}
					}
				}

				data.Functions = append(data.Functions, function)
			case "AWS::Serverless::GraphQLApi":
				function := defaultFunction
				function.InvocationType = FaaSInvocationTypeGraphQL
				function.Location = FaaSLocationRegion
				function.SourceFilePath = config.Path
				data.Functions = append(data.Functions, function)
			case "AWS::CloudFront::Distribution":
				properties, err := JsonResolveMap(resource, []string{"Properties"})
				if err != nil {
					continue
				}

				distributionConfig, err := JsonResolveMap(properties, []string{"DistributionConfig"})
				if err != nil {
					continue
				}

				allCacheBehaviors := make([]interface{}, 0)

				defaultCacheBehavior, err := JsonResolveMap(distributionConfig, []string{"DefaultCacheBehavior"})
				if err == nil {
					allCacheBehaviors = append(allCacheBehaviors, defaultCacheBehavior)
				}

				cacheBehaviors, err := JsonResolveArray(distributionConfig, []string{"CacheBehaviors"})
				if err == nil {
					allCacheBehaviors = append(allCacheBehaviors, cacheBehaviors...)
				}

				for _, cacheBehavior := range allCacheBehaviors {
					lambdaFunctionAssociations, err := JsonResolveArray(cacheBehavior, []string{"LambdaFunctionAssociations"})
					if err == nil {
						numEdgeLambdaFunctions += len(lambdaFunctionAssociations)
					}

					functionAssociations, err := JsonResolveArray(cacheBehavior, []string{"FunctionAssociations"})
					if err == nil {
						for _, _ = range functionAssociations {
							function := defaultFunction
							function.InvocationType = FaaSInvocationTypeHTTP
							function.Location = FaaSLocationEdge
							function.SourceFilePath = config.Path
							data.Functions = append(data.Functions, function)
						}
					}

				}
			default:
			}
		}
	}

	// TODO: figure out what to do if numEdgeLambdaFunctions > len(prelimFunctions)
	if numEdgeLambdaFunctions <= len(prelimFunctions) {
		for index := 0; index < numEdgeLambdaFunctions; index++ {
			prelimFunctions[index].Location = FaaSLocationEdge
		}
		data.Functions = append(data.Functions, prelimFunctions...)
	}
}

func scanAzureResourceManager(data *RepositoryData, files []TextFile) {
	armJsonFiles, err := FilterTextFiles(
		files,
		"**/*.json",
	)
	if err != nil {
		return
	}

	armBicepFiles, err := FilterTextFiles(
		files,
		"**/*.bicep",
	)
	if err != nil {
		return
	}

	if len(armBicepFiles)+len(armJsonFiles) == 0 {
		return
	}

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformAzure,
		Framework:      FaaSFrameworkAzureResourceManager,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationRegion,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	for _, armJsonFile := range armJsonFiles {
		armJson, err := LoadJsonFromBytes([]byte(armJsonFile.Content))
		if err != nil {
			continue
		}

		schema, err := JsonResolveString(armJson, []string{"$schema"})
		if err != nil {
			continue
		}

		switch schema {
		case "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
			"https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
			"https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
			"https://schema.management.azure.com/schemas/2019-08-01/managementGroupDeploymentTemplate.json#",
			"https://schema.management.azure.com/schemas/2019-08-01/tenantDeploymentTemplate.json#":
			// pass
		default:
			continue
		}

		resources := make([]interface{}, 0)

		resourcesMap, err1 := JsonResolveMap(armJson, []string{"resources"})
		if err1 == nil {
			for _, resource := range resourcesMap {
				resources = append(resources, resource)
			}
		}

		resourcesArray, err2 := JsonResolveArray(armJson, []string{"resources"})
		if err2 == nil {
			for _, resource := range resourcesArray {
				resources = append(resources, resource)
			}
		}

		if err1 != nil && err2 != nil {
			continue
		}

		data.UsedPlatforms[FaaSPlatformAzure] = true
		data.UsedFrameworks[FaaSFrameworkAzureResourceManager] = true

		for _, resource := range resources {
			resourceType, err := JsonResolveString(resource, []string{"type"})
			if err != nil {
				continue
			}

			if resourceType != "Microsoft.Web/sites" {
				continue
			}

			resourceKind, err := JsonResolveString(resource, []string{"kind"})
			if err != nil {
				continue
			}

			function := defaultFunction
			function.SourceFilePath = armJsonFile.Path

			// https://github.com/Azure/app-service-linux-docs/blob/master/Things_You_Should_Know/kind_property.md
			switch resourceKind {
			case "app", "hyperV", "app,container,windows",
				"app,linux", "app,linux,container":
				function.InvocationType = FaaSInvocationTypeHTTP
				data.Functions = append(data.Functions, function)
			default:
				function.InvocationType = FaaSInvocationTypeUnknown
				data.Functions = append(data.Functions, function)
			}
		}
	}

	for _, armBicepFile := range armBicepFiles {
		numFunctions := strings.Count(armBicepFile.Content, "Microsoft.Web/sites@20")

		data.UsedPlatforms[FaaSPlatformAzure] = true
		data.UsedFrameworks[FaaSFrameworkAzureResourceManager] = true

		for index := 0; index < numFunctions; index++ {
			function := defaultFunction
			function.SourceFilePath = armBicepFile.Path
			function.InvocationType = FaaSInvocationTypeUnknown
			data.Functions = append(data.Functions, function)
		}
	}
}

func scanGCPCloudDeploymentManager(data *RepositoryData, files []TextFile) {
	ValidResourceTypes := []string{
		"spanner.v1.instance",
		"compute.beta.image",
		"runtimeconfig.v1beta1.config",
		"compute.alpha.regionBackendService",
		"compute.v1.regionInstanceGroupManager",
		"compute.alpha.backendService",
		"compute.beta.router",
		"compute.alpha.targetTcpProxy",
		"compute.beta.subnetwork",
		"compute.v1.externalVpnGateway",
		"deploymentmanager.v2beta.typeProvider",
		"compute.alpha.targetVpnGateway",
		"compute.beta.disk",
		"compute.beta.regionHealthCheck",
		"logging.v2beta1.metric",
		"compute.v1.firewall",
		"compute.v1.router",
		"replicapool.v1beta1.pool",
		"iam.v1.serviceAccounts.key",
		"compute.v1.regionBackendService",
		"compute.v1.instanceGroupManager",
		"bigquery.v2.dataset",
		"compute.v1.sslCertificate",
		"compute.v1.disk",
		"sqladmin.v1beta4.instance",
		"compute.v1.image",
		"compute.alpha.subnetwork",
		"runtimeconfig.v1beta1.waiter",
		"compute.beta.regionInstanceGroup",
		"compute.alpha.disk",
		"compute.alpha.instanceGroup",
		"pubsub.v1.topic",
		"compute.beta.globalForwardingRule",
		"compute.beta.sslCertificate",
		"compute.v1.targetInstance",
		"compute.v1.healthCheck",
		"compute.v1.subnetwork",
		"compute.alpha.xpnResource",
		"compute.alpha.regionHealthCheck",
		"compute.alpha.backendBucket",
		"compute.alpha.route",
		"sqladmin.v1beta4.sslCert",
		"runtimeconfig.v1beta1.variable",
		"storage.v1.object",
		"compute.beta.backendBucket",
		"compute.beta.vpnTunnel",
		"compute.beta.route",
		"compute.v1.autoscaler",
		"compute.alpha.httpsHealthCheck",
		"compute.alpha.address",
		"compute.beta.vpnGateway",
		"compute.v1.targetTcpProxy",
		"compute.v1.targetSslProxy",
		"storage.v1.defaultObjectAccessControl",
		"sqladmin.v1beta4.database",
		"compute.beta.autoscaler",
		"bigtableadmin.v2.instance",
		"storage.v1.bucket",
		"compute.beta.xpnHost",
		"storage.v1.objectAccessControl",
		"resourceviews.v1beta2.zoneView",
		"compute.alpha.urlMap",
		"compute.alpha.globalAddress",
		"compute.v1.route",
		"compute.v1.httpHealthCheck",
		"dns.v1.managedZone",
		"dataproc.v1.cluster",
		"compute.v1.regionHealthCheck",
		"compute.beta.forwardingRule",
		"compute.beta.regionAutoscaler",
		"container.v1.cluster",
		"bigquery.v2.table",
		"compute.beta.urlMap",
		"compute.beta.httpsHealthCheck",
		"logging.v2.metric",
		"compute.alpha.regionInstanceGroup",
		"compute.v1.vpnTunnel",
		"compute.beta.regionBackendService",
		"compute.beta.interconnectAttachment",
		"compute.alpha.targetPool",
		"compute.v1.interconnectAttachment",
		"compute.alpha.regionAutoscaler",
		"compute.alpha.targetHttpProxy",
		"compute.v1.instanceGroup",
		"compute.v1.urlMap",
		"compute.alpha.forwardingRule",
		"compute.alpha.vpnGateway",
		"cloudresourcemanager.v1.project",
		"compute.alpha.targetInstance",
		"compute.beta.httpHealthCheck",
		"compute.v1.regionAutoscaler",
		"compute.alpha.healthCheck",
		"compute.alpha.targetSslProxy",
		"compute.beta.targetVpnGateway",
		"logging.v2.sink",
		"compute.beta.network",
		"compute.v1.httpsHealthCheck",
		"compute.v1.vpnGateway",
		"sqladmin.v1beta4.user",
		"compute.alpha.vpnTunnel",
		"compute.alpha.interconnectAttachment",
		"compute.v1.forwardingRule",
		"compute.beta.instanceGroupManager",
		"compute.v1.regionInstanceGroup",
		"compute.beta.targetSslProxy",
		"compute.alpha.router",
		"compute.v1.targetPool",
		"compute.beta.targetHttpProxy",
		"compute.alpha.targetHttpsProxy",
		"compute.alpha.globalForwardingRule",
		"compute.alpha.firewall",
		"compute.alpha.instanceTemplate",
		"compute.beta.firewall",
		"compute.v1.targetHttpProxy",
		"compute.beta.xpnResource",
		"compute.v1.address",
		"compute.v1.globalAddress",
		"logging.v2beta1.sink",
		"compute.beta.regionInstanceGroupManager",
		"compute.beta.targetTcpProxy",
		"compute.alpha.externalVpnGateway",
		"compute.v1.targetHttpsProxy",
		"compute.v1.globalForwardingRule",
		"compute.v1.instanceTemplate",
		"compute.v1.backendService",
		"compute.beta.backendService",
		"pubsub.v1.subscription",
		"compute.beta.instance",
		"compute.alpha.autoscaler",
		"compute.v1.network",
		"compute.beta.globalAddress",
		"compute.alpha.regionInstanceGroupManager",
		"bigtableadmin.v2.instance.table",
		"compute.alpha.network",
		"compute.v1.targetVpnGateway",
		"compute.alpha.xpnHost",
		"compute.beta.targetPool",
		"compute.beta.address",
		"storage.v1.bucketAccessControl",
		"compute.beta.healthCheck",
		"compute.alpha.sslCertificate",
		"compute.beta.externalVpnGateway",
		"compute.alpha.httpHealthCheck",
		"compute.beta.instanceTemplate",
		"compute.alpha.instanceGroupManager",
		"compute.alpha.image",
		"compute.beta.targetHttpsProxy",
		"compute.beta.targetInstance",
		"appengine.v1.version",
		"container.v1.nodePool",
		"compute.beta.instanceGroup",
		"compute.v1.instance",
		"iam.v1.serviceAccount",
		"compute.alpha.instance",
	}

	usesGCPCloudDeploymentManager := false

	dmConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, dmConfigFile := range dmConfigFiles {
		dmConfigJson, err := LoadJsonFromBytes([]byte(dmConfigFile.Content))
		if err != nil {
			continue
		}

		resources, err := JsonResolveArray(dmConfigJson, []string{"resources"})
		if err != nil {
			continue
		}

		for _, resource := range resources {
			resourceType, err := JsonResolveString(resource, []string{"type"})
			if err != nil {
				continue
			}

			if strings.HasPrefix(resourceType, "gcp-types/") ||
				strings.HasSuffix(resourceType, ".jinja") ||
				strings.HasSuffix(resourceType, ".py") ||
				slices.Contains(ValidResourceTypes, resourceType) {
				usesGCPCloudDeploymentManager = true
			}
		}
	}

	if !usesGCPCloudDeploymentManager {
		return
	}

	data.UsedFrameworks[FaaSFrameworkGCPCloudDeploymentManager] = true
	data.UsedPlatforms[FaaSPlatformGCP] = true

	dmFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
		"**/*.jinja", "**/*.py",
	)
	if err != nil {
		return
	}

	for _, dmFile := range dmFiles {
		numFunctions := strings.Count(dmFile.Content, "gcp-types/cloudfunctions-v1:projects.locations.functions") +
			strings.Count(dmFile.Content, "gcp-types/cloudfunctions-v2beta:projects.locations.functions")

		for index := 0; index < numFunctions; index++ {
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformGCP,
				Framework:      FaaSFrameworkGCPCloudDeploymentManager,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: dmFile.Path,
				SourceFileLine: -1,
			})
		}
	}
}

func scanAlibabaResourceOrchestrationService(data *RepositoryData, files []TextFile) {
	configs, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
		"**/*.json",
	)
	if err != nil {
		return
	}

	if len(configs) == 0 {
		return
	}

	for _, config := range configs {
		var configJson interface{}
		switch config.Extension {
		case ".yaml", ".yml":
			configJsons := LoadJsonsFromYamlBytes([]byte(config.Content))
			if len(configJsons) != 1 {
				continue
			}
			configJson = configJsons[0]
		case ".json":
			configJson, err = LoadJsonFromBytes([]byte(config.Content))
			if err != nil {
				continue
			}
		default:
			continue
		}

		resources, err := JsonResolveMap(configJson, []string{"Resources"})
		if err != nil {
			continue
		}

		for _, resource := range resources {
			resourceType, err := JsonResolveString(resource, []string{"Type"})
			if err != nil {
				continue
			}

			data.UsedPlatforms[FaaSPlatformAWS] = true
			data.UsedFrameworks[FaaSFrameworkAWSCloudFormationAndSAM] = true

			function := RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformAlibaba,
				Framework:      FaaSFrameworkAlibabaResourceOrchestrationService,
				InvocationType: FaaSInvocationTypeUnknown, // set below
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: config.Path,
				SourceFileLine: -1,
			}

			switch resourceType {
			case "ALIYUN::FC::Function":
				function.InvocationType = FaaSInvocationTypeUnknown
				data.Functions = append(data.Functions, function)
			default:
			}
		}
	}
}

func scanFnProject(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, 0)
	allDependencies = append(allDependencies, data.Dependencies...)
	allDependencies = append(allDependencies, data.DevDependencies...)

	if !slices.Contains(allDependencies, "@fnproject/fdk") {
		return
	}

	data.UsedFrameworks[FaaSFrameworkFnProject] = true
	data.UsedPlatforms[FaaSPlatformFnProject] = true

	jsFiles, err := FilterTextFiles(
		files,
		"**/*.js", "**/*.jsx", "**/*.mjs",
	)
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !strings.Contains(jsFile.Content, "@fnproject/fdk") {
			continue
		}

		if !strings.Contains(jsFile.Content, ".handle(") {
			continue
		}

		data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
			Name:           "",
			Platform:       FaaSPlatformFnProject,
			Framework:      FaaSFrameworkFnProject,
			InvocationType: FaaSInvocationTypeHTTP,
			Location:       FaaSLocationRegion,
			TimeoutSeconds: -1,
			SourceFilePath: jsFile.Path,
			SourceFileLine: -1,
		})
	}
}

func scanNuclio(data *RepositoryData, files []TextFile) {
	nuclioConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, nuclioConfigFile := range nuclioConfigFiles {
		nuclioConfigJsons := LoadJsonsFromYamlBytes([]byte(nuclioConfigFile.Content))
		for _, nuclioConfigJson := range nuclioConfigJsons {
			apiVersion, err := JsonResolveString(nuclioConfigJson, []string{"apiVersion"})
			if err != nil {
				continue
			}

			if !strings.HasPrefix(apiVersion, "nuclio.io/") {
				continue
			}

			kind, err := JsonResolveString(nuclioConfigJson, []string{"kind"})
			if err != nil {
				continue
			}

			if kind != "NuclioFunction" {
				continue
			}

			data.UsedPlatforms[FaaSPlatformNuclio] = true
			data.UsedFrameworks[FaaSFrameworkNuclio] = true
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformNuclio,
				Framework:      FaaSFrameworkNuclio,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: nuclioConfigFile.Path,
				SourceFileLine: -1,
			})
		}
	}
}

func scanOpenWhisk(data *RepositoryData, files []TextFile) {
	openWhiskConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, openWhiskConfigFile := range openWhiskConfigFiles {
		openWhiskConfigJsons := LoadJsonsFromYamlBytes([]byte(openWhiskConfigFile.Content))
		if len(openWhiskConfigJsons) != 1 {
			continue
		}
		openWhiskConfigJson := openWhiskConfigJsons[0]

		openWhiskPackages, err := JsonResolveMap(openWhiskConfigJson, []string{"packages"})
		if err != nil {
			continue
		}

		for _, openWhiskPackage := range openWhiskPackages {
			openWhiskPackageActions, err := JsonResolveMap(openWhiskPackage, []string{"actions"})
			if err != nil {
				continue
			}

			data.UsedPlatforms[FaaSPlatformOpenWhisk] = true
			data.UsedFrameworks[FaaSFrameworkOpenWhisk] = true

			numFunctions := len(openWhiskPackageActions)
			for index := 0; index < numFunctions; index++ {
				data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
					Name:           "",
					Platform:       FaaSPlatformOpenWhisk,
					Framework:      FaaSFrameworkOpenWhisk,
					InvocationType: FaaSInvocationTypeUnknown,
					Location:       FaaSLocationRegion,
					TimeoutSeconds: -1,
					SourceFilePath: openWhiskConfigFile.Path,
					SourceFileLine: -1,
				})
			}
		}
	}
}

func scanFission(data *RepositoryData, files []TextFile) {
	fissionConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, fissionConfigFile := range fissionConfigFiles {
		fissionConfigJsons := LoadJsonsFromYamlBytes([]byte(fissionConfigFile.Content))
		for _, fissionConfigJson := range fissionConfigJsons {
			apiVersion, err := JsonResolveString(fissionConfigJson, []string{"apiVersion"})
			if err != nil {
				continue
			}

			if !strings.HasPrefix(apiVersion, "fission.io/") {
				continue
			}

			kind, err := JsonResolveString(fissionConfigJson, []string{"kind"})
			if err != nil {
				continue
			}

			if kind != "Function" {
				continue
			}

			data.UsedPlatforms[FaaSPlatformFission] = true
			data.UsedFrameworks[FaaSFrameworkFission] = true
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformFission,
				Framework:      FaaSFrameworkFission,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: fissionConfigFile.Path,
				SourceFileLine: -1,
			})
		}
	}
}

func scanKubeless(data *RepositoryData, files []TextFile) {
	kubelessConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, kubelessConfigFile := range kubelessConfigFiles {
		kubelessConfigJsons := LoadJsonsFromYamlBytes([]byte(kubelessConfigFile.Content))
		for _, kubelessConfigJson := range kubelessConfigJsons {
			apiVersion, err := JsonResolveString(kubelessConfigJson, []string{"apiVersion"})
			if err != nil {
				continue
			}

			if !strings.HasPrefix(apiVersion, "kubeless.io/") {
				continue
			}

			kind, err := JsonResolveString(kubelessConfigJson, []string{"kind"})
			if err != nil {
				continue
			}

			if kind != "Function" {
				continue
			}

			data.UsedPlatforms[FaaSPlatformKubeless] = true
			data.UsedFrameworks[FaaSFrameworkKubeless] = true
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformKubeless,
				Framework:      FaaSFrameworkKubeless,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: kubelessConfigFile.Path,
				SourceFileLine: -1,
			})
		}
	}
}

func scanKnative(data *RepositoryData, files []TextFile) {
	knativeConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, knativeConfigFile := range knativeConfigFiles {
		knativeConfigJsons := LoadJsonsFromYamlBytes([]byte(knativeConfigFile.Content))
		for _, knativeConfigJson := range knativeConfigJsons {
			apiVersion, err := JsonResolveString(knativeConfigJson, []string{"apiVersion"})
			if err != nil {
				continue
			}

			if !strings.HasPrefix(apiVersion, "serving.knative.dev/") {
				continue
			}

			kind, err := JsonResolveString(knativeConfigJson, []string{"kind"})
			if err != nil {
				continue
			}

			if kind != "Service" {
				continue
			}

			data.UsedPlatforms[FaaSPlatformKnative] = true
			data.UsedFrameworks[FaaSFrameworkKnative] = true
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformKnative,
				Framework:      FaaSFrameworkKnative,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: knativeConfigFile.Path,
				SourceFileLine: -1,
			})
		}
	}
}

func scanFirebase(data *RepositoryData, files []TextFile) {
	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformFirebase,
		Framework:      FaaSFrameworkFirebase,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationRegion,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	firebaseConfigFiles, err := FilterTextFiles(files, "**/firebase.json")
	if err == nil && len(firebaseConfigFiles) > 0 {
		data.UsedPlatforms[FaaSPlatformFirebase] = true
		data.UsedFrameworks[FaaSFrameworkFirebase] = true

		for _, firebaseConfigFile := range firebaseConfigFiles {
			firebaseConfigJson, err := LoadJsonFromBytes([]byte(firebaseConfigFile.Content))
			if err != nil {
				continue
			}

			functions, err := JsonResolveArray(firebaseConfigJson, []string{"functions"})
			if err != nil {
				continue
			}

			for _, _ = range functions {
				function := defaultFunction
				function.SourceFilePath = firebaseConfigFile.Path
				function.InvocationType = FaaSInvocationTypeUnknown
				data.Functions = append(data.Functions, function)
			}
		}
	}

	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.Dependencies...)
	allDependencies = append(allDependencies, data.DevDependencies...)

	if !slices.Contains(allDependencies, "firebase-functions") {
		return
	}

	data.UsedPlatforms[FaaSPlatformFirebase] = true
	data.UsedFrameworks[FaaSFrameworkFirebase] = true

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !strings.Contains(jsFile.Content, "firebase-functions") {
			continue
		}

		function := defaultFunction
		function.SourceFilePath = jsFile.Path

		// 2nd gen
		switch {
		case strings.Contains(jsFile.Content, "https") && (strings.Contains(jsFile.Content, "onRequest(") ||
			strings.Contains(jsFile.Content, "onCall(")):
			function.InvocationType = FaaSInvocationTypeHTTP
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "pubsub") && strings.Contains(jsFile.Content, "onMessagePublished("):
			function.InvocationType = FaaSInvocationTypeTopic
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "eventarc") && strings.Contains(jsFile.Content, "onCustomEventPublished("):
			function.InvocationType = FaaSInvocationTypeTopic
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "scheduler") && strings.Contains(jsFile.Content, "onSchedule("):
			function.InvocationType = FaaSInvocationTypeSchedule
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "database") && (strings.Contains(jsFile.Content, "onValueCreated(") ||
			strings.Contains(jsFile.Content, "onValueDeleted(") ||
			strings.Contains(jsFile.Content, "onValueUpdated(") ||
			strings.Contains(jsFile.Content, "onValueWritten(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "firestore") && (strings.Contains(jsFile.Content, "onDocumentCreated(") ||
			strings.Contains(jsFile.Content, "onDocumentDeleted(") ||
			strings.Contains(jsFile.Content, "onDocumentUpdated(") ||
			strings.Contains(jsFile.Content, "onDocumentWritten(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "storage") && (strings.Contains(jsFile.Content, "onObjectArchived(") ||
			strings.Contains(jsFile.Content, "onObjectDeleted(") ||
			strings.Contains(jsFile.Content, "onObjectFinalized(") ||
			strings.Contains(jsFile.Content, "onObjectMetadataUpdated(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "tasks") && strings.Contains(jsFile.Content, "onTaskDispatched("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "alerts") && strings.Contains(jsFile.Content, "onAlertPublished("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "remoteConfig") && strings.Contains(jsFile.Content, "onConfigUpdated("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "identity") && (strings.Contains(jsFile.Content, "beforeOperation(") ||
			strings.Contains(jsFile.Content, "beforeUserCreated(") ||
			strings.Contains(jsFile.Content, "beforeUserSignedIn(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		default:
		}

		// 1st gen
		switch {
		case strings.Contains(jsFile.Content, "https") && (strings.Contains(jsFile.Content, "onRequest(") ||
			strings.Contains(jsFile.Content, "onCall(")):
			function.InvocationType = FaaSInvocationTypeHTTP
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "pubsub") && strings.Contains(jsFile.Content, "schedule("):
			function.InvocationType = FaaSInvocationTypeSchedule
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "pubsub") && strings.Contains(jsFile.Content, "topic("):
			function.InvocationType = FaaSInvocationTypeTopic
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "database") && (strings.Contains(jsFile.Content, "onCreate(") ||
			strings.Contains(jsFile.Content, "onDelete(") ||
			strings.Contains(jsFile.Content, "onUpdate(") ||
			strings.Contains(jsFile.Content, "onWrite(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "firestore") && (strings.Contains(jsFile.Content, "onCreate(") ||
			strings.Contains(jsFile.Content, "onDelete(") ||
			strings.Contains(jsFile.Content, "onUpdate(") ||
			strings.Contains(jsFile.Content, "onWrite(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "storage") && (strings.Contains(jsFile.Content, "bucket(") ||
			strings.Contains(jsFile.Content, "object(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "tasks") && strings.Contains(jsFile.Content, "taskQueue("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "remoteConfig") && strings.Contains(jsFile.Content, "onUpdate("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "analytics") && strings.Contains(jsFile.Content, "onLog("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "auth") && (strings.Contains(jsFile.Content, "beforeCreate(") ||
			strings.Contains(jsFile.Content, "beforeSignIn(") ||
			strings.Contains(jsFile.Content, "onCreate(") ||
			strings.Contains(jsFile.Content, "onDelete(")):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		default:
		}
	}
}

func scanFastly(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.DevDependencies...)
	allDependencies = append(allDependencies, data.Dependencies...)
	if slices.Contains(allDependencies, "@fastly/js-compute") {
		data.UsedPlatforms[FaaSPlatformFastly] = true
		data.UsedFrameworks[FaaSFrameworkFastly] = true
	}

	fastlyConfigFiles, err := FilterTextFiles(files, "**/fastly.toml")
	if err != nil {
		return
	}

	if len(fastlyConfigFiles) == 0 {
		return
	}

	data.UsedPlatforms[FaaSPlatformFastly] = true
	data.UsedFrameworks[FaaSFrameworkFastly] = true

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}
	for _, jsFile := range jsFiles {
		if strings.Contains(jsFile.Content, "addEventListener('fetch'") ||
			strings.Contains(jsFile.Content, "addEventListener(\"fetch\"") {
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformFastly,
				Framework:      FaaSFrameworkFastly,
				InvocationType: FaaSInvocationTypeHTTP,
				Location:       FaaSLocationEdge,
				TimeoutSeconds: -1,
				SourceFilePath: jsFile.Path,
				SourceFileLine: -1,
			})
		}
	}

}

func scanCloudflare(data *RepositoryData, files []TextFile) {
	wranglerConfigFiles, err := FilterTextFiles(files, "**/wrangler.toml")
	if err != nil || len(wranglerConfigFiles) == 0 {
		return
	}

	data.UsedPlatforms[FaaSPlatformCloudflare] = true
	data.UsedFrameworks[FaaSFrameworkWrangler] = true

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformCloudflare,
		Framework:      FaaSFrameworkWrangler,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationEdge,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if (strings.Contains(jsFile.Content, "export default") &&
			strings.Contains(jsFile.Content, "fetch(")) ||
			strings.Contains(jsFile.Content, "addEventListener('fetch'") ||
			strings.Contains(jsFile.Content, "addEventListener(\"fetch\"") {
			function := defaultFunction
			function.InvocationType = FaaSInvocationTypeHTTP
			function.SourceFilePath = jsFile.Path
			data.Functions = append(data.Functions, function)
		}

		if (strings.Contains(jsFile.Content, "export default") &&
			strings.Contains(jsFile.Content, "queue(")) ||
			strings.Contains(jsFile.Content, "addEventListener('queue'") ||
			strings.Contains(jsFile.Content, "addEventListener(\"queue\"") {
			function := defaultFunction
			function.InvocationType = FaaSInvocationTypeQueue
			function.SourceFilePath = jsFile.Path
			data.Functions = append(data.Functions, function)
		}

		if (strings.Contains(jsFile.Content, "export default") &&
			strings.Contains(jsFile.Content, "scheduled(")) ||
			strings.Contains(jsFile.Content, "addEventListener('scheduled'") ||
			strings.Contains(jsFile.Content, "addEventListener(\"scheduled\"") {
			function := defaultFunction
			function.InvocationType = FaaSInvocationTypeSchedule
			function.SourceFilePath = jsFile.Path
			data.Functions = append(data.Functions, function)
		}

		if strings.Contains(jsFile.Content, "export") && (strings.Contains(jsFile.Content, "onRequest") ||
			strings.Contains(jsFile.Content, "onRequestGet") ||
			strings.Contains(jsFile.Content, "onRequestPost") ||
			strings.Contains(jsFile.Content, "onRequestPatch") ||
			strings.Contains(jsFile.Content, "onRequestPut") ||
			strings.Contains(jsFile.Content, "onRequestDelete") ||
			strings.Contains(jsFile.Content, "onRequestHead") ||
			strings.Contains(jsFile.Content, "onRequestOptions")) {
			function := defaultFunction
			function.InvocationType = FaaSInvocationTypeHTTP
			function.SourceFilePath = jsFile.Path
			data.Functions = append(data.Functions, function)
		}
	}
}

func scanTencent(data *RepositoryData, files []TextFile) {
	scfConfigFiles, err := FilterTextFiles(files, "**/serverless.yml", "**/serverless.yaml")
	if err != nil {
		return
	}

	for _, scfConfigFile := range scfConfigFiles {
		scfConfigJsons := LoadJsonsFromYamlBytes([]byte(scfConfigFile.Content))
		if len(scfConfigJsons) != 1 {
			continue
		}
		scfConfigJson := scfConfigJsons[0]

		component, err := JsonResolveString(scfConfigJson, []string{"component"})
		if err != nil || component != "scf" {
			continue
		}

		if _, err := JsonResolveString(scfConfigJson, []string{"inputs", "src"}); err != nil {
			continue
		}

		if _, err := JsonResolveString(scfConfigJson, []string{"inputs", "handler"}); err != nil {
			continue
		}

		if _, err := JsonResolveString(scfConfigJson, []string{"inputs", "runtime"}); err != nil {
			continue
		}

		if _, err := JsonResolveArray(scfConfigJson, []string{"inputs", "events"}); err != nil {
			continue
		}

		data.UsedPlatforms[FaaSPlatformTencent] = true
		data.UsedFrameworks[FaaSFrameworkServerlessCloudFramework] = true
		data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
			Name:           "",
			Platform:       FaaSPlatformTencent,
			Framework:      FaaSFrameworkServerlessCloudFramework,
			InvocationType: FaaSInvocationTypeUnknown,
			Location:       FaaSLocationRegion,
			TimeoutSeconds: -1,
			SourceFilePath: scfConfigFile.Path,
			SourceFileLine: -1,
		})
	}
}

func scanOpenFaaS(data *RepositoryData, files []TextFile) {
	openFaaSConfigFiles, err := FilterTextFiles(
		files,
		"**/*.yaml", "**/*.yml",
	)
	if err != nil {
		return
	}

	for _, openFaaSConfigFile := range openFaaSConfigFiles {
		openFaaSConfigJsons := LoadJsonsFromYamlBytes([]byte(openFaaSConfigFile.Content))
		for _, openFaaSConfigJson := range openFaaSConfigJsons {
			apiVersion, err := JsonResolveString(openFaaSConfigJson, []string{"apiVersion"})
			if err != nil {
				continue
			}

			if !strings.HasPrefix(apiVersion, "openfaas.com/") {
				continue
			}

			kind, err := JsonResolveString(openFaaSConfigJson, []string{"kind"})
			if err != nil {
				continue
			}

			if kind != "Function" {
				continue
			}

			data.UsedPlatforms[FaaSPlatformOpenFaaS] = true
			data.UsedFrameworks[FaaSFrameworkOpenFaaS] = true
			data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
				Name:           "",
				Platform:       FaaSPlatformOpenFaaS,
				Framework:      FaaSFrameworkOpenFaaS,
				InvocationType: FaaSInvocationTypeUnknown,
				Location:       FaaSLocationRegion,
				TimeoutSeconds: -1,
				SourceFilePath: openFaaSConfigFile.Path,
				SourceFileLine: -1,
			})
		}
	}
}

func scanDigitalOcean(data *RepositoryData, files []TextFile) {
	digitalOceanConfigFiles, err := FilterTextFiles(files, "**/project.yml", "**/project.yaml")
	if err != nil {
		return
	}

	for _, digitalOceanConfigFile := range digitalOceanConfigFiles {
		digitalOceanConfigs := LoadJsonsFromYamlBytes([]byte(digitalOceanConfigFile.Content))
		if len(digitalOceanConfigs) != 1 {
			continue
		}
		digitalOceanConfig := digitalOceanConfigs[0]

		digitalOceanPackages, err := JsonResolveArray(digitalOceanConfig, []string{"packages"})
		if err != nil {
			continue
		}

		if len(digitalOceanPackages) != 1 {
			continue
		}

		data.UsedPlatforms[FaaSPlatformDigitalOcean] = true
		data.UsedFrameworks[FaaSFrameworkDigitalOcean] = true

		defaultFunction := RepositoryFaaSFunctionData{
			Name:           "",
			Platform:       FaaSPlatformDigitalOcean,
			Framework:      FaaSFrameworkDigitalOcean,
			InvocationType: FaaSInvocationTypeUnknown, // set below
			Location:       FaaSLocationRegion,
			TimeoutSeconds: -1,
			SourceFilePath: digitalOceanConfigFile.Path,
			SourceFileLine: -1,
		}

		for _, digitalOceanPackage := range digitalOceanPackages {
			allDigitalOceanFunctions := make([]interface{}, 0)

			digitalOceanFunctions, err := JsonResolveArray(digitalOceanPackage, []string{"functions"})
			if err == nil {
				allDigitalOceanFunctions = append(allDigitalOceanFunctions, digitalOceanFunctions...)
			}

			digitalOceanActions, err := JsonResolveArray(digitalOceanPackage, []string{"actions"})
			if err == nil {
				allDigitalOceanFunctions = append(allDigitalOceanFunctions, digitalOceanActions...)
			}

			for _, digitalOceanFunction := range allDigitalOceanFunctions {
				function := defaultFunction

				triggers, err := JsonResolveArray(digitalOceanFunction, []string{"triggers"})
				if err == nil {
					for _, trigger := range triggers {
						sourceType, err := JsonResolveString(trigger, []string{"sourceType"})
						if err != nil {
							continue
						}

						if sourceType == "scheduler" {
							function.InvocationType = FaaSInvocationTypeSchedule
						}

					}
				}

				data.Functions = append(data.Functions, function)
			}
		}
	}
}

func scanAzureFunctionsFramework(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.DevDependencies...)
	allDependencies = append(allDependencies, data.Dependencies...)

	if slices.Contains(allDependencies, "@azure/functions") ||
		slices.Contains(allDependencies, "azure-functions-core-tools") {
		data.UsedPlatforms[FaaSPlatformAzure] = true
		data.UsedFrameworks[FaaSFrameworkAzureFunctions] = true
	}

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformAzure,
		Framework:      FaaSFrameworkAzureFunctions,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationRegion,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	// Check for Version 3

	affConfigFiles, err := FilterTextFiles(files, "**/function.json")
	if err == nil {
		for _, affConfigFile := range affConfigFiles {
			affConfig, err := LoadJsonFromBytes([]byte(affConfigFile.Content))
			if err != nil {
				continue
			}

			affBindings, err := JsonResolveArray(affConfig, []string{"bindings"})
			if err != nil {
				continue
			}

			if len(affBindings) == 0 {
				continue
			}

			data.UsedPlatforms[FaaSPlatformAzure] = true
			data.UsedFrameworks[FaaSFrameworkAzureFunctions] = true

			function := defaultFunction
			function.SourceFilePath = affConfigFile.Path

			for _, affBinding := range affBindings {
				affBindingDirection, err := JsonResolveString(affBinding, []string{"direction"})
				if err != nil {
					continue
				}

				if affBindingDirection != "in" {
					continue
				}

				affBindingType, err := JsonResolveString(affBinding, []string{"type"})
				if err != nil {
					continue
				}

				switch affBindingType {
				case "httpTrigger":
					function.InvocationType = FaaSInvocationTypeHTTP
				case "timerTrigger":
					function.InvocationType = FaaSInvocationTypeSchedule
				case "queueTrigger":
					function.InvocationType = FaaSInvocationTypeQueue
				case "serviceBusTrigger":
					if _, err := JsonResolveString(affBinding, []string{"queueName"}); err == nil {
						function.InvocationType = FaaSInvocationTypeQueue
					}
					if _, err := JsonResolveString(affBinding, []string{"topicName"}); err == nil {
						function.InvocationType = FaaSInvocationTypeTopic
					}
				}
			}

			data.Functions = append(data.Functions, function)
		}
	}

	// Check for Version 4

	if !slices.Contains(allDependencies, "@azure/functions") {
		return
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !strings.Contains(jsFile.Content, "@azure/functions") {
			continue
		}

		function := defaultFunction
		function.SourceFilePath = jsFile.Path

		switch {
		case strings.Contains(jsFile.Content, ".http("):
			function.InvocationType = FaaSInvocationTypeHTTP
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, ".timer("):
			function.InvocationType = FaaSInvocationTypeSchedule
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, ".storageQueue("),
			strings.Contains(jsFile.Content, ".serviceBusQueue(") &&
				strings.Contains(jsFile.Content, "queueName"):
			function.InvocationType = FaaSInvocationTypeQueue
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, ".serviceBusQueue(") &&
			strings.Contains(jsFile.Content, "topicName"):
			function.InvocationType = FaaSInvocationTypeTopic
			data.Functions = append(data.Functions, function)
		default:
		}

	}
}

func scanGCPFunctionsFramework(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.DevDependencies...)
	allDependencies = append(allDependencies, data.Dependencies...)

	if !slices.Contains(allDependencies, "@google-cloud/functions-framework") {
		return
	}

	data.UsedPlatforms[FaaSPlatformGCP] = true
	data.UsedFrameworks[FaaSFrameworkGCPFunctions] = true

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformGCP,
		Framework:      FaaSFrameworkGCPFunctions,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationRegion,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !strings.Contains(jsFile.Content, "@google-cloud/functions-framework") {
			continue
		}

		function := defaultFunction
		function.SourceFilePath = jsFile.Path

		switch {
		case strings.Contains(jsFile.Content, ".http("):
			function.InvocationType = FaaSInvocationTypeHTTP
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, ".cloudEvent("):
			function.InvocationType = FaaSInvocationTypeUnknown
			data.Functions = append(data.Functions, function)
		default:
		}
	}
}

func scanDurableFunctionsFramework(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.DevDependencies...)
	allDependencies = append(allDependencies, data.Dependencies...)

	if !slices.Contains(allDependencies, "durable-functions") {
		return
	}

	data.UsedPlatforms[FaaSPlatformAzure] = true
	data.UsedFrameworks[FaaSFrameworkAzureDurableFunctions] = true

	defaultFunction := RepositoryFaaSFunctionData{
		Name:           "",
		Platform:       FaaSPlatformAzure,
		Framework:      FaaSFrameworkAzureDurableFunctions,
		InvocationType: FaaSInvocationTypeUnknown, // set below
		Location:       FaaSLocationRegion,
		TimeoutSeconds: -1,
		SourceFilePath: "", // set below
		SourceFileLine: -1,
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !strings.Contains(jsFile.Content, "durable-functions") {
			continue
		}

		function := defaultFunction
		function.SourceFilePath = jsFile.Path

		switch {
		// Check for Version 3
		case strings.Contains(jsFile.Content, ".orchestration("),
			strings.Contains(jsFile.Content, ".activity("):
			fallthrough
		// Check for Version 2
		case strings.Contains(jsFile.Content, ".orchestrator("):
			function.InvocationType = FaaSInvocationTypeOther
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "client.http("):
			function.InvocationType = FaaSInvocationTypeHTTP
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "client.timer("):
			function.InvocationType = FaaSInvocationTypeSchedule
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "client.storageQueue("),
			strings.Contains(jsFile.Content, "client.serviceBusQueue(") &&
				strings.Contains(jsFile.Content, "queueName"):
			function.InvocationType = FaaSInvocationTypeQueue
			data.Functions = append(data.Functions, function)
		case strings.Contains(jsFile.Content, "client.serviceBusQueue(") &&
			strings.Contains(jsFile.Content, "topicName"):
			function.InvocationType = FaaSInvocationTypeTopic
			data.Functions = append(data.Functions, function)
		default:
		}
	}
}

func scanAlexaSkillsKit(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.DevDependencies...)
	allDependencies = append(allDependencies, data.Dependencies...)

	if !slices.Contains(allDependencies, "ask-sdk-core") {
		return
	}

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !(strings.Contains(jsFile.Content, "ask-sdk-core") &&
			strings.Contains(jsFile.Content, ".SkillBuilders") &&
			strings.Contains(jsFile.Content, ".lambda(")) {
			continue
		}

		data.UsedPlatforms[FaaSPlatformAWS] = true
		data.UsedFrameworks[FaaSFrameworkAlexaSkillsKit] = true

		data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
			Name:           "",
			Platform:       FaaSPlatformAWS,
			Framework:      FaaSFrameworkAlexaSkillsKit,
			InvocationType: FaaSInvocationTypeOther,
			Location:       FaaSLocationRegion,
			TimeoutSeconds: -1,
			SourceFilePath: jsFile.Path,
			SourceFileLine: -1,
		})
	}
}

func scanHono(data *RepositoryData, files []TextFile) {
	allDependencies := make([]string, len(data.Dependencies)+len(data.DevDependencies))
	allDependencies = append(allDependencies, data.DevDependencies...)
	allDependencies = append(allDependencies, data.Dependencies...)

	if !slices.Contains(allDependencies, "hono") {
		return
	}

	data.UsedPlatforms[FaaSPlatformCloudflare] = true
	data.UsedPlatforms[FaaSPlatformFastly] = true
	data.UsedPlatforms[FaaSPlatformAWS] = true
	data.UsedFrameworks[FaaSFrameworkHono] = true

	jsFiles, err := FilterTextFiles(files, "**/*.js", "**/*.jsx", "**/*.mjs")
	if err != nil {
		return
	}

	for _, jsFile := range jsFiles {
		if !(strings.Contains(jsFile.Content, "\"hono\"") ||
			strings.Contains(jsFile.Content, "'hono'")) {
			continue
		}

		if !strings.Contains(jsFile.Content, "new Hono") {
			continue
		}

		data.Functions = append(data.Functions, RepositoryFaaSFunctionData{
			Name:           "",
			Platform:       FaaSPlatformUnknown,
			Framework:      FaaSFrameworkHono,
			InvocationType: FaaSInvocationTypeHTTP,
			Location:       FaaSLocationUnknown,
			TimeoutSeconds: -1,
			SourceFilePath: jsFile.Path,
			SourceFileLine: -1,
		})
	}
}

func SaveRepositoryData(repositoryData RepositoryData, outPath string) error {
	repositoryDataBytes, err := json.Marshal(repositoryData)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, repositoryDataBytes, 0644); err != nil {
		return err
	}

	return nil
}

func LoadRepositoryData(inPath string) (RepositoryData, error) {
	repositoryDataBytes, err := os.ReadFile(inPath)
	if err != nil {
		return RepositoryData{}, err
	}

	var repositoryData RepositoryData
	err = json.Unmarshal(repositoryDataBytes, &repositoryData)
	if err != nil {
		return RepositoryData{}, err
	}

	return repositoryData, nil
}

func AggregateRepositoryData(
	httpClient *http.Client,
	repositoryId RepositoryId,
	repositoriesDirectory string,
	repositoryInfosDirectory string,
	repositoryIssuesCommitsAndContributorsDirectory string,
	excludeDirectories []string,
) (RepositoryData, error) {
	var result RepositoryData

	// Load info
	repositoryInfo, err := LoadRepositoryInfo(path.Join(
		repositoryInfosDirectory,
		fmt.Sprintf("%d.json", repositoryId),
	))
	if err != nil {
		return RepositoryData{}, err
	}

	// Load metadata
	ricc, err := LoadRepositoryIssuesCommitsAndContributors(path.Join(
		repositoryIssuesCommitsAndContributorsDirectory,
		fmt.Sprintf("%d.json", repositoryId),
	))
	if err != nil {
		return RepositoryData{}, err
	}

	result.RepositoryId = repositoryId
	result.Url = repositoryInfo.GetHTMLURL()
	result.Name = repositoryInfo.GetName()
	result.Description = repositoryInfo.GetDescription()
	result.License = repositoryInfo.GetLicense().GetName()
	result.Topics = repositoryInfo.Topics
	result.Stars = repositoryInfo.GetStargazersCount()
	result.Watchers = repositoryInfo.GetWatchersCount()
	result.Forks = repositoryInfo.GetForksCount()

	result.Size = repositoryInfo.GetSize()
	result.Archived = repositoryInfo.GetArchived()
	result.Forked = repositoryInfo.GetFork()

	result.CreatedAt = repositoryInfo.GetCreatedAt().Time
	result.PushedAt = repositoryInfo.GetPushedAt().Time

	result.NumContributors = ricc.NumContributors
	result.NumIssues = ricc.NumOpenIssues + ricc.NumClosedIssues
	result.NumOpenIssues = ricc.NumOpenIssues
	result.NumClosedIssues = ricc.NumClosedIssues

	result.NumCommits = ricc.NumCommits

	result.FirstCommitAt, result.LastCommitAt = getFirstAndLastCommitDate(ricc.CommitsHeadAndTail)
	result.ActiveDays = durationInDays(result.LastCommitAt.Sub(result.FirstCommitAt))

	result.FirstHumanCommitAt, result.LastHumanCommitAt = getFirstAndLastCommitDate(
		filterHumanCommits(ricc.CommitsHeadAndTail),
	)
	result.ActiveHumanDays = durationInDays(result.LastHumanCommitAt.Sub(result.FirstHumanCommitAt))

	repositoryFiles, err := LoadTextFiles(path.Join(
		repositoriesDirectory,
		fmt.Sprintf("%d", repositoryId),
		fmt.Sprintf("**/*"),
	), excludeDirectories)
	if err != nil {
		return RepositoryData{}, err
	}

	packageJsonFiles, err := FilterTextFiles(
		repositoryFiles,
		path.Join(
			repositoriesDirectory,
			fmt.Sprintf("%d", repositoryId),
			fmt.Sprintf("**/package.json"),
		),
	)
	if err != nil {
		fmt.Printf("error finding package json files: %v\n", err)
		return RepositoryData{}, err
	}

	result.Packages = make([]RepositoryPackageData, 0)
	for _, packageJsonFile := range packageJsonFiles {
		var packageData RepositoryPackageData
		packageData.RootPath = path.Dir(packageJsonFile.Path)

		scanPackageJson(&packageData, packageJsonFile)
		scanFaaSRuntimeDependencies(&packageData)

		applicationFiles, err := FilterTextFiles(
			repositoryFiles,
			path.Join(
				packageData.RootPath,
				"**/*",
			),
		)
		if err != nil {
			fmt.Printf("error finding application files: %v\n", err)
			continue
		}

		//packageData.NumFilesByExtension = countFilesByExtension(applicationFiles)
		//packageData.LinesOfTextByExtension = countNumberOfLinesByExtension(applicationFiles)
		scanFaaSHandlers(&packageData, applicationFiles)
		scanPackageJsons(&packageData, applicationFiles)

		result.Packages = append(result.Packages, packageData)
	}
	result.NumPackages = len(result.Packages)

	result.Dependencies = make([]string, 0)
	result.DevDependencies = make([]string, 0)
	result.FaaSRuntimeDependencies = make([]string, 0)

	result.NumPublishedToNPM = 0
	result.NumFaaSHandlers = 0
	for _, packageData := range result.Packages {
		if packageData.PublishedToNPM {
			result.NumPublishedToNPM += 1
		}
		result.NumFaaSHandlers += packageData.NumFaaSHandlers
		result.Dependencies = append(result.Dependencies, packageData.Dependencies...)
		result.DevDependencies = append(result.DevDependencies, packageData.DevDependencies...)
		result.FaaSRuntimeDependencies = append(result.FaaSRuntimeDependencies, packageData.FaaSRuntimeDependencies...)
	}

	result.Dependencies = UniqueSliceElements(result.Dependencies)
	result.DevDependencies = UniqueSliceElements(result.DevDependencies)
	result.FaaSRuntimeDependencies = UniqueSliceElements(result.FaaSRuntimeDependencies)

	result.NumFaaSRuntimeDependencies = len(result.FaaSRuntimeDependencies)

	result.Complexity = extractComplexity(repositoryFiles)

	result.UsedFrameworks = make(map[FaaSFramework]bool)
	result.UsedPlatforms = make(map[FaaSPlatform]bool)
	result.Functions = make([]RepositoryFaaSFunctionData, 0)

	scanVercel(&result, repositoryFiles)
	scanNetlify(&result, repositoryFiles)
	scanServerless(&result, repositoryFiles)
	scanNitric(&result, repositoryFiles)
	scanArchitect(&result, repositoryFiles)
	scanAWSCDKAndSST(&result, repositoryFiles)
	scanTerraform(&result, repositoryFiles)
	scanPulumi(&result, repositoryFiles)
	scanAWSCloudFormationAndSAM(&result, repositoryFiles)
	scanAzureResourceManager(&result, repositoryFiles)
	scanGCPCloudDeploymentManager(&result, repositoryFiles)
	scanAlibabaResourceOrchestrationService(&result, repositoryFiles)
	scanFnProject(&result, repositoryFiles)
	scanNuclio(&result, repositoryFiles)
	scanOpenWhisk(&result, repositoryFiles)
	scanFission(&result, repositoryFiles)
	scanKubeless(&result, repositoryFiles)
	scanKnative(&result, repositoryFiles)
	scanFirebase(&result, repositoryFiles)
	scanFastly(&result, repositoryFiles)
	scanCloudflare(&result, repositoryFiles)
	scanTencent(&result, repositoryFiles)
	scanOpenFaaS(&result, repositoryFiles)
	scanDigitalOcean(&result, repositoryFiles)
	scanAzureFunctionsFramework(&result, repositoryFiles)
	scanGCPFunctionsFramework(&result, repositoryFiles)
	scanDurableFunctionsFramework(&result, repositoryFiles)
	scanAlexaSkillsKit(&result, repositoryFiles)
	scanHono(&result, repositoryFiles)

	result.NumFunctions = len(result.Functions)

	return result, nil
}

func AggregateRepositoriesData(
	numWorkers int,
	repositoryIds []RepositoryId,
	repositoriesDirectory string,
	repositoryInfosDirectory string,
	repositoryIssuesCommitsAndContributorsDirectory string,
	excludeDirectories []string,
	outDirectory string,
) {
	ProcessInParallel(
		repositoryIds,
		func(repositoryId RepositoryId, httpClient *http.Client, _ GitHubClient) (RepositoryData, bool) {
			repositoryData, err := AggregateRepositoryData(
				httpClient,
				repositoryId,
				repositoriesDirectory,
				repositoryInfosDirectory,
				repositoryIssuesCommitsAndContributorsDirectory,
				excludeDirectories,
			)
			if err != nil {
				fmt.Printf("error aggregating repository data: %v\n", err)
				return RepositoryData{}, false
			}
			return repositoryData, true
		},
		func(repositoryData RepositoryData) {
			if err := SaveRepositoryData(
				repositoryData,
				path.Join(outDirectory, fmt.Sprintf("%d.json", repositoryData.RepositoryId)),
			); err != nil {
				fmt.Printf("error saving repository data: %v\n", err)
			}
		},
		numWorkers,
		1,
		10_000,
		10_000,
	)
}
