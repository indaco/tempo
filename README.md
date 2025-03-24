<h1 align="center">
  <code>tempo</code>
</h1>
<h2 align="center" style="font-size: 1.5em;">
  A lightweight CLI for managing assets and scaffolding components in templ-based projects.
</h2>
<p align="center">
    <a href="https://github.com/indaco/tempo/actions/workflows/ci.yml" target="_blank">
      <img src="https://github.com/indaco/tempo/actions/workflows/ci.yml/badge.svg" alt="CI" />
    </a>
    <a href="https://coveralls.io/github/indaco/tempo?branch=main" target="_blank">
      <img src="https://coveralls.io/repos/github/indaco/tempo/badge.svg?branch=main" alt="Coverage Status" />
    </a>
    <a href="https://goreportcard.com/report/github.com/indaco/tempo" target="_blank">
      <img src="https://goreportcard.com/badge/github.com/indaco/tempo" alt="go report card" />
    </a>
   <a href="https://badge.fury.io/gh/indaco%2Ftempo" target="_blank">
    <img src="https://badge.fury.io/gh/indaco%2Ftempo.svg" alt="version" height="18" />
   </a>
    <a href="https://pkg.go.dev/github.com/indaco/tempo/" target="_blank">
      <img src="https://pkg.go.dev/badge/github.com/indaco/tempo/.svg" alt="go reference" />
    </a>
    <a href="https://github.com/indaco/tempo/blob/main/LICENSE" target="_blank">
      <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square&logo=none" alt="license" />
    </a>
    <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/" target="_blank">
      <img src="https://www.jetify.com/img/devbox/shield_moon.svg" alt="Built with Devbox" />
    </a>
</p>

`tempo` is a lightweight CLI for managing assets and scaffolding components in <a href="https://templ.guide" target="_blank">templ</a>-based projects. Inspired by the Italian word for **"time"**, it streamlines CSS & JS workflows while preserving a smooth developer experience.

![tempo](demo.gif)

## 📖 Table of Contents

- [✨ Features](#-features)
- [💡 Motivation](#-motivation)
- [💻 Installation](#-installation)
- [🚀 Usage](#-usage)
- [🛠️ CLI Commands & Options](#️-cli-commands--options)
- [⚡ Using `tempo sync` as a Standalone Command](#-using-tempo-sync-as-a-standalone-command)
- [⚙️ Configuration](#️-configuration)
- [🏗️ Templates & Actions](#️-templates--actions)
- [🔌 Extending Tempo](#-extending-tempo)
- [🤝 Contributing](#-contributing)
- [🆓 License](#-license)

## ✨ Features

- **Automated CSS & JS management** – Keeps styles and scripts in separate files while seamlessly injecting them into `.templ` components.
- **Structured asset workflow** – Ensures a clean, maintainable approach to managing CSS and JS within `templ` projects.
- **Fast, lightweight component scaffolding** – Quickly generate components and variants with predefined templates and actions.
- **Extensible templating system** – Supports custom function providers (local or remote) to enhance `tempo`'s capabilities.

## 💡 Motivation

While building a UI component library in Golang with `templ`, I deliberately chose to use plain CSS and vanilla JavaScript. This decision introduced two key challenges:

### The Problem

1. **Managing CSS & JS assets** – While `templ` excels at Go/HTML templating, it lacks a structured approach for handling standalone styles and scripts. Although you can write CSS and JS directly within `.templ` files, this comes at the cost of losing native tooling benefits such as syntax highlighting, formatting, and autocompletion. As a result, maintaining styles and scripts efficiently while keeping them separate from `.templ` files required a better workflow.
2. **Scaffolding new components** – Every component followed the same folder structure, but manually copying files and folders was inefficient. I initially used [Plop.js](https://plopjs.com/), but it required setting up a full Node.js project.

### The Solution

`tempo` solves both problems natively in Go, eliminating the need for Node.js while providing a **structured, opinionated workflow** for component and asset management.

#### ✨ Preserving the Developer Experience

With `tempo`, CSS and JS files remain untouched during development, allowing developers to continue using their preferred tools:

- **Linters & formatters** (e.g., Prettier, ESLint, Stylelint)
- **IDE features** like syntax highlighting, error detection, and inline code suggestions (e.g., in _VSCode_).
- **Existing workflows** remain intact—ensuring a familiar, efficient experience.

When you run `tempo`, **CSS and JS files are injected into `.templ` components automatically**, but the original source files remain unchanged. This approach **preserves developer productivity** while enabling seamless integration into templ-based projects.

## 💻 Installation

### go install

Requires Go v1.23+

```bash
go install github.com/indaco/tempo/cmd/tempo@latest
```

### Manually

Download the pre-compiled binaries from the [releases page](https://github.com/indaco/tempo/releases) and place it in a directory available in your system's _PATH_.

## 🚀 Usage

To start using `tempo`, initialize your project and define your components. Below are the key steps:

**1. Initialize tempo**

```bash
tempo init
```

Generates a `tempo.yaml` configuration file. Customize it to fit your project. See the [Configuration](#️-configuration) section for details.

**2. Define a Component or Variant**

```bash
tempo component define
tempo variant define
```

Generates templates for _components/variants_ inside `.tempo-files/templates/` along with an action JSON file inside `.tempo-files/actions/`. See the [Templates & Actions](#️-templates--actions) section for details.

**3. Create a Component**

```bash
tempo component new --name button
```

Creates a new component:

- `assets/button/` → Holds CSS and JS files.
- `components/button/` → Contains _.templ_ and _.go_ file

**4. Sync Assets with Components**

```bash
tempo sync
```

Injects CSS & JS into _.templ_ files using guard markers.

🔹 Example mapping:

```bash
assets/button/css/button.css          → components/button/css/button.templ
assets/button/css/variant/outline.css → components/button/css/variant/outline.templ
```

## 🛠️ CLI Commands & Options

```bash
NAME:
   tempo - A lightweight CLI for managing assets and scaffolding components in templ-based projects.

USAGE:
   tempo <subcommand> [options] [arguments]

VERSION:
   v0.2.1

DESCRIPTION:
   tempo simplifies asset management in templ-based projects, providing a seamless workflow for
   handling CSS and JS files. It automatically extracts and injects styles and scripts into .templ
   components while preserving the original source files, ensuring a smooth developer experience.
   Additionally, it offers a lightweight scaffolding system to quickly generate component and variant
   templates with predefined structures.

COMMANDS:
   init       Initialize a Tempo project
   component  Define component templates and generate instances from them
   variant    Define variant templates and generate instances from them
   register   Register is used to extend tempo
   sync       Process & sync asset files into templ components
   help       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### init

Initialize a tempo project by generating a `tempo.yaml` configuration file. This file allows you to **customize how components and assets are generated and managed**.

For a full list of configuration options, See the [Configuration](#️-configuration) section for details.

### component

Define reusable component templates and generate instances from them.

#### Define a Component

```bash
tempo component define
```

Creates a scaffold for defining a UI component, including template files and an action JSON file.

These definitions are later used by `tempo component new` to generate real components.

<details>
<summary><strong>Flags</strong> (<code>tempo component define</code>) </summary>
<dl>
  <dt><code>--force</code></dt>
  <dd>Force overwriting if already exists (<em>default: false</em>)</dd>

  <dt><code>--dry-run</code></dt>
  <dd>Preview actions without making changes (<em>default: false</em>)</dd>

  <dt><code>--js</code></dt>
  <dd>Whether or not JS is needed for the component (<em>default: false</em>)</dd>
</dl>
</details>

#### Generate a Component

```bash
tempo component new --name button
```

Uses the templates and actions created with `tempo component define` to generate a real component inside _assets/_ and _components/_.

Example: Running `tempo component new --name dropdown` will generate:

- `assets/dropdown/` _(CSS & JS files)_
- `components/dropdown/` _(with .templ and .go files)_

<details>
<summary><strong>Flags</strong> (<code>tempo component new</code>)</summary>

<dl>
  <dt><code>--package</code> (<code>-p</code>) <em>value</em></dt>
  <dd>The Go package name where components will be generated (<em>default: components</em>)</dd>

  <dt><code>--assets</code> (<code>-a</code>) <em>value</em></dt>
  <dd>The directory where asset files (e.g., CSS, JS) will be generated (<em>default: assets</em>)</dd>

  <dt><code>--name</code> (<code>-n</code>) <em>value</em></dt>
  <dd>Name of the component</dd>

  <dt><code>--js</code></dt>
  <dd>Whether or not JS is needed for the component (<em>default: false</em>)</dd>

  <dt><code>--force</code></dt>
  <dd>Force overwriting if the component already exists (<em>default: false</em>)</dd>

  <dt><code>--dry-run</code></dt>
  <dd>Preview actions without making changes (<em>default: false</em>)</dd>
</dl>
</details>

### variant

Define component variant templates and create instances based on them.

#### Define a Variant

```bash
tempo variant define
```

Creates a scaffold for defining a **variant** of an existing component (e.g., an **outline** button variant), including template files and an action JSON file.

These definitions are later used by `tempo variant new` to generate real component variants.

<details>
<summary><strong>Flags</strong> (<code>tempo variant define</code>) </summary>

<dl>
  <dt><code>--force</code></dt>
  <dd>Force overwriting if the variant definition already exists (<em>default: false</em>)</dd>
  <dt><code>--dry-run</code></dt>
  <dd>Preview actions without making changes (<em>default: false</em>)</dd>
</dl>

</details>

#### Generate a Variant

```bash
tempo variant new --name outline --component button
```

Uses the templates and actions created with `tempo variant define` to generate a real instance of a variant inside the corresponding component folder.

Example: Running `tempo variant new --name outline --component` button will generate:

- `components/button/variant/outline.templ`
- `assets/button/css/variant/outline.css`

<details>
<summary><strong>Flags</strong> (<code>tempo variant new</code>)</summary>

<dl>
  <dt><code>--package</code> (<code>-p</code>) <em>value</em></dt>
  <dd>The Go package name where components will be generated (<em>default: components</em>)</dd>

  <dt><code>--assets</code> (<code>-a</code>) <em>value</em></dt>
  <dd>The directory where asset files (e.g., CSS, JS) will be generated (<em>default: assets</em>)</dd>

  <dt><code>--name</code> (<code>-n</code>) <em>value</em> <strong>(required)</strong></dt>
  <dd>The name of the variant being generated</dd>

  <dt><code>--component</code> (<code>-c</code>) <em>value</em> <strong>(required)</strong></dt>
  <dd>Name of the component or entity to which this variant belongs</dd>

  <dt><code>--force</code></dt>
  <dd>Force overwriting if the variant already exists (<em>default: false</em>)</dd>

  <dt><code>--dry-run</code></dt>
  <dd>Preview actions without making changes (<em>default: false</em>)</dd>
</dl>

</details>

### register

Extend `tempo` with the `register` command.

```bash
tempo register -h
```

#### Register a function provider

Enhance `tempo`’s templating capabilities by adding custom function providers. You can register a provider from either a **local Go module** or a **remote Git repository**.

```bash
tempo register functions --name sprig --url https://github.com/indaco/tempo-provider-sprig.git
```

<details>
<summary><strong>Flags</strong> (<code>tempo register functions</code>)</summary>
<dl>
  <dt><code>--name</code> (<code>-n</code>) <em>value</em></dt>
  <dd>Name for the function provider</dd>

  <dt><code>--url</code> (<code>-u</code>) <em>value</em></dt>
  <dd>Repository URL</dd>

  <dt><code>--path</code> (<code>-p</code>) <em>value</em></dt>
  <dd>Path to a local go module provider</dd>
</dl>
</details>

### sync

Automatically sync CSS and JS assets with `.templ` components.

```bash
tempo sync
```

The sync command scans the `input` folder for CSS and JS files, then injects their content into the corresponding `.templ` files in the `output` folder:

- **Extracts** CSS and JS from source files.
- **Injects** them into `.templ` files inside sections marked by _guard markers_.
- **Keeps components up to date** without manual copying.

Whenever you update your CSS or JS files, simply run `tempo sync` to propagate the changes.

<details>
<summary><strong>Flags</strong></summary>
<dl>
  <dt><code>--input</code> <em>value</em></dt>
  <dd>The directory containing asset files (e.g., CSS, JS) to be processed (<em>default: assets</em>)</dd>

  <dt><code>--output</code> <em>value</em></dt>
  <dd>The directory containing the `.templ` component files where assets will be injected (<em>default: components</em>)</dd>

  <dt><code>--exclude</code> <em>value</em></dt>
  <dd>Subfolder (relative to input directory) to exclude from the processing</dd>

  <dt><code>--workers</code> <em>value</em></dt>
  <dd>Number of concurrent workers processing asset files (<em>default: numCPUs * 2</em>)</dd>

  <dt><code>--prod</code></dt>
  <dd>Enable production mode, minifying the injected content (<em>default: false</em>)</dd>

  <dt><code>--force</code></dt>
  <dd>Force processing of all files, ignoring modification timestamps (<em>default: false</em>)</dd>

  <dt><code>--summary</code> (<code>-s</code>) <em>value</em></dt>
  <dd>Specify the summary format: <code>text</code>, <code>json</code>, or <code>none</code> (<em>default: "text"</em>)</dd>

  <dt><code>--verbose</code></dt>
  <dd>Show detailed information in the summary report (<em>default: false</em>)</dd>

  <dt><code>--report-file</code> (<code>--rf</code>) <em>value</em></dt>
  <dd>Export the summary to a JSON file</dd>
</dl>
</details>

> [!TIP]
> **Live Reload:** If you're using a live reload tool like [air](https://github.com/air-verse/air), [templier](https://github.com/romshark/templier), or [watchexec](https://github.com/watchexec/watchexec), pass `--summary none` to reduce unnecessary output.

> [!NOTE]
> **Want to use tempo without scaffolding?**
> Check out [Using tempo sync as a Standalone Command](#-using-tempo-sync-as-a-standalone-command).

## ⚡ Using tempo sync as a Standalone Command

The `sync` command can be used **independently** of the scaffolding flow (`define`, `new`, `register`). All it requires is:

- A valid `tempo.yaml` file.
- A project folder structure matching the paths defined in `tempo.yaml`.
- **Guard markers** in your `.templ` files to specify where assets should be injected.

This makes `tempo sync` a great option for developers who simply want to **synchronize CSS and JS assets** with their `.templ` files, without using the full scaffolding features.

### 💥 Use Cases for Standalone sync

- Already have components but want automated asset injection.
- Only need asset handling without using the full scaffolding flow.
- Maintain existing folder structures while benefiting from tempo’s synchronization features.

### 🛠️ Example Standalone Workflow

Start by creating a Go module using the `go mod init` [command](https://go.dev/ref/mod#go-mod-init).

1. **Initialize `tempo.yaml`:**

   ```bash
   tempo init
   ```

2. **Set up your folders:**

   ```bash
   .
   ├── assets/
   │   └── button/
   │       └── button.css
   └── components/
       └── button/
           └── button.templ
   ```

3. **Prepare .templ file with guard markers:**

   ```templ
   package button

   var buttonCSSHandle = templ.NewOnceHandle()

   templ Button() {
     @buttonCSSHandle.Once() {
       <style type="text/css">
       /* [tempo] BEGIN - Do not edit! This section is auto-generated. */
       /* [tempo] END */
       </style>
   }
   ```

4. **Run the sync command:**

   ```bash
   tempo sync
   ```

5. **Result:**

CSS and JS are injected into the corresponding `.templ` files, replacing the content between the guard markers.

### 📋 Guard Markers Explained

`tempo sync` uses guard markers in `.templ` files to locate where CSS and JS should be injected.

By default, it looks for the following markers:

```templ
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */
```

Only the text inside square brackets (`[tempo]`) can be customized in your `tempo.yaml` file under the `templates.guard_marker` setting.

> [!IMPORTANT]
> If no guard markers are present, `tempo sync` will **skip** the file, ensuring only intended sections are updated.

## ⚙️ Configuration

<details>
<summary><strong>📄 View Default <code>tempo.yaml</code> Configuration</strong></summary>

    # The root folder for Tempo files
    tempo_root: .tempo-files

    app:
      # The name of the Go module being worked on.
      go_module: <FROM_GO.MOD>

      # The Go package name where components will be organized and generated.
      go_package: components

      # The directory where asset files (CSS, JS) will be generated.
      assets_dir: assets

      # Indicates whether JavaScript is required for the component.
      # with_js: false

      # The name of the CSS layer to associate with component styles.
      # css_layer: components

    processor:
      # Number of concurrent workers (numCPUs * 2).
      workers: 4

      # Summary format: long, compact, json, none.
      summary_format: compact

    templates:
      # A text placeholder or sentinel used in template files to mark auto-generated content.
      guard_marker: tempo

      # File extensions used for template files.
      extensions:
        - .gotxt
        - .gotmpl
        - .tpl

</details>

### Configuration Options

#### Project Structure

| Key          | Default        | Description                                                 |
| :----------- | :------------- | :---------------------------------------------------------- |
| `tempo_root` | `.tempo-files` | Root folder where `tempo` stores its templates and actions. |

#### Application-Specific Settings

| Key              | Default      | Description                                               |
| :--------------- | :----------- | :-------------------------------------------------------- |
| `app.go_module`  |              | Name of the Go module being worked on read from `go.mod`. |
| `app.go_package` | `components` | Go package name where components will be generated.       |
| `app.assets_dir` | `assets`     | Directory where asset files (CSS, JS) will be generated.  |
| `app.with_js`    | `false`      | Whether JavaScript is required for the component.         |
| `app.css_layer`  | `components` | CSS layer name to associate with component styles.        |

#### Files Processing Options

| Key                        | Default | Description                                        |
| :------------------------- | :------ | :------------------------------------------------- |
| `processor.workers`        | `4`     | Number of concurrent workers.                      |
| `processor.summary_format` | `long`  | Summary format: `long`, `compact`, `json`, `none`. |

#### Templates Options

| Key                            | Default                     | Description                                                                                 |
| :----------------------------- | :-------------------------- | :------------------------------------------------------------------------------------------ |
| `templates.extensions`         | `.gotxt`, `.gotmpl`, `.tpl` | File extensions used for template files.                                                    |
| `templates.guard_marker`       | `tempo`                     | Placeholder used in templates to mark auto-generated content.                               |
| `templates.user_data`          | `nil`                       | Custom variables (flat or nested) passed to templates under `.UserData`.                    |
| `templates.function_providers` | `[]` _(empty)_              | A list of external function providers that can be loaded from a local path or a remote URL. |

##### Function Provider Format

Each function provider entry follows this format:

```yaml
templates:
  function_providers:
    - name: default
      type: path
      value: ./providers/default
    - name: custom
      type: url
      value: https://github.com/user/custom-provider.git
```

##### Fields

| Field   | Type     | Description                                                           |
| :------ | :------- | :-------------------------------------------------------------------- |
| `name`  | `string` | (Optional) The name of the function provider.                         |
| `type`  | `string` | Specifies if the provider is a local path or a remote url (Git repo). |
| `value` | `string` | The actual path or URL where the provider is located.                 |

#### 🛠️ Example Use Case

If your project follows a custom structure, you can update `tempo.yaml` like this:

```yaml
app:
  go_module: myproject
  package: ui/components
  assets_dir: static/assets
  with_js: true
  css_layer: my-layer
```

This ensures:

- Assets are stored in `static/assets`
- Components are generated under `ui/components`
- JavaScript support is enabled
- CSS styles are associated with the `my-layer` layer

## 🏗️ Templates & Actions

📌 The **`tempo component define`** and **`tempo variant define`** commands generate:

- **Templates** stored in `.tempo-files/templates/`, using Go’s `text/template`.
- **Actions** defined in `.tempo-files/actions/`, specifying _file and folder creation_ in JSON format.

### Templates

Templates define the **structure of components and variants**. They use **Go’s `text/template`** engine along with custom template functions provided by Tempo.

#### 📌 Default Template Functions

Tempo provides a set of built-in helper functions:

| Function           | Description                                                     |
| :----------------- | :-------------------------------------------------------------- |
| `goPackageName`    | Converts a string into a valid Go package name.                 |
| `goExportedName`   | Converts a string into a valid **exported** Go function name.   |
| `goUnexportedName` | Converts a string into a valid **unexported** Go function name. |
| `normalizePath`    | Normalizes a path string.                                       |
| `isEmpty`          | Checks if a string is empty.                                    |

#### 📌 Built-in Template Variables

Tempo automatically provides a set of **predefined variables** that can be used inside templates. These variables come from the configuration and CLI context during execution.

| Variable        | Description                                                           |
| :-------------- | :-------------------------------------------------------------------- |
| `TemplatesDir`  | The root directory containing template files.                         |
| `ActionsDir`    | The root directory containing actions files.                          |
| `GoModule`      | The name of the Go module being worked on.                            |
| `GoPackage`     | The Go package name where components will be organized and generated. |
| `ComponentName` | The name of the component being generated.                            |
| `VariantName`   | The name of the variant being generated (if applicable).              |
| `AssetsDir`     | The directory where asset files (CSS, JS) will be generated.          |
| `WithJs`        | Whether JavaScript is required for the component (true/false).        |
| `CssLayer`      | The CSS layer name associated with component styles.                  |
| `GuardMarker`   | Placeholder used in _templ_ files to mark auto-generated sections.    |

#### 📌 Extending Template Functions

Tempo supports external function providers, allowing you to integrate additional helper functions into your templates.

See the full guide in [Extending Tempo](#-extending-tempo).

### Actions

Actions define **how templates should be processed**. They are stored in `.tempo-files/actions/` as JSON files.

Example **component action file (component.json)**:

```json
[
  {
    "item": "file",
    "templateFile": "component/templ/component.templ.gotxt",
    "path": "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/{{ .ComponentName | goPackageName }}.templ"
  },
  {
    "item": "file",
    "templateFile": "component/assets/css/base.css.gotxt",
    "path": "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/base.css"
  },
  {
    "item": "folder",
    "source": "component/assets/css/themes",
    "destination": "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/themes"
  }
]
```

Each object defines a templating action:

| Key            | Type          | Description                                        |
| :------------- | :------------ | :------------------------------------------------- |
| `item`         | `file/folder` | Whether the action is creating a file or a folder. |
| `templateFile` | `string`      | Path to the template file (_only for file items_). |
| `path`         | `string`      | Output path for the generated file.                |
| `source`       | `string`      | Source folder (_only for folder items_).           |
| `destination`  | `string`      | Destination folder for copied content.             |
| `skipIfExists` | `bool`        | Whether to skip the file if it already exists.     |
| `force`        | `bool`        | Whether to overwrite existing files.               |

> [!NOTE]
> When `item` is `folder`, all files inside the source folder will be processed and copied to the destination folder.

**Variant actions (variant.json) follow the same structure** but target a specific component’s variant subfolder.

## 🔌 Extending tempo

`tempo` provides two ways to extend how templates work:

1. **Custom template variables** — enrich templates with additional context using `user_data`.
2. **Custom template functions** — bring in new helpers via external function providers.

### 🔧 Adding Custom Template Variables

You can pass additional variables to your templates using the `user_data` section in the `tempo.yaml` config file. These variables are accessible inside templates using `.UserData`.

**Example: tempo.yaml**

```yaml
# ....
templates:
  user_data:
    author: Jane Doe
    year: 2025
    config:
      option1: true
      option2: false
```

**Basic Access (dot notation)**

```go
Author: {{ .UserData.author }}
Year: {{ .UserData.year }}
```

**Nested Access**

You can use either:

- `index`(built-in from [text/template](https://pkg.go.dev/text/template))
- `lookup`(provided by _tempo_, supports dot notation)

```go
{{ index (index .UserData "config") "option1" }}
{{ lookup .UserData "config.option1" }}
```

### 📦 Using External Function Providers

You can add external function providers in two ways:

**1. Via the tempo.yaml configuration file:**

```yaml
templates:
  function_providers:
    - name: sprig
      type: url
      value: https://github.com/indaco/tempo-provider-sprig.git
```

**2. Via the CLI using the tempo register functions command:**

- Register a GitHub repository provider:

  ```bash
  tempo register functions --name sprig --url https://github.com/indaco/tempo-provider-sprig.git
  ```

- Register a local provider from a directory:

  ```bash
  tempo register functions --name myprovider --path /path/to/myprovider_module
  ```

### 🛠 Implementing a Custom Template Function Provider

To create a custom function provider, implement the `TemplateFuncProvider` interface from the `github.com/indaco/tempo-api`.

See [tempo api](https://github.com/indaco/tempo-api) for full details.

**Example: Creating a Custom Function Provider**

```go
package myprovider

import (
    "text/template"
    "github.com/indaco/tempo-api/templatefuncs"
)

// MyProvider implements the TemplateFuncProvider interface
type MyProvider struct{}

// GetFunctions returns a map of function names to implementations
func (p *MyProvider) GetFunctions() template.FuncMap {
    return template.FuncMap{
        "myFunc": func() string { return "Hello from myFunc!" },
    }
}

// Provider instance
var Provider templatefuncs.TemplateFuncProvider = &MyProvider{}
```

Once your provider is implemented:

- If published as a Git repository, register it using:

  ```bash
  tempo register functions --name myprovider --url https://github.com/user/myprovider_module.git
  ```

- If stored locally, register it using:

  ```bash
  tempo register functions --name myprovider --path /path/to/myprovider_module
  ```

### 📦 Available Function Providers

A pre-built function provider for `Masterminds/sprig` is [available](https://github.com/indaco/tempo-provider-sprig.git) for convenience:

```bash
tempo register functions --name sprig --url https://github.com/indaco/tempo-provider-sprig.git
```

This allows you to access [Sprig functions](https://github.com/Masterminds/sprig) within Tempo templates.

## 🤝 Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setting up the development tools.

## 🆓 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
