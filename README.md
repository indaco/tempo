<h1 align="center">
  <code>tempo</code>
</h1>
<h2 align="center" style="font-size: 1.5em;">
  A simple CLI for scaffolding components and managing assets in templ-based projects.
</h2>
<p align="center">
    <a href="https://github.com/indaco/tempo/actions/workflows/ci.yml" target="_blank">
      <img src="https://github.com/indaco/tempo/actions/workflows/ci.yml/badge.svg" alt="CI" />
    </a>
    <a href="https://coveralls.io/github/indaco/tempo?branch=main">
        <img
            src="https://coveralls.io/repos/github/indaco/tempo/badge.svg?branch=main"
            alt="Coverage Status"
        />
    </a>
    <a href="https://goreportcard.com/report/github.com/indaco/tempo" target="_blank">
        <img src="https://goreportcard.com/badge/github.com/indaco/tempo" alt="go report card" />
    </a>
    <a href="https://pkg.go.dev/github.com/indaco/tempo/" target="_blank">
        <img src="https://pkg.go.dev/badge/github.com/indaco/tempo/.svg" alt="go reference" />
    </a>
     <a href="https://github.com/indaco/tempo/blob/main/LICENSE" target="_blank">
        <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square&logo=none" alt="license" />
    </a>
    <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/">
      <img
          src="https://www.jetify.com/img/devbox/shield_moon.svg"
          alt="Built with Devbox"
      />
    </a>
</p>

`tempo` is a simple CLI for accelerating scaffolding and asset management in <a href="https://templ.guide" target="_blank">templ</a>-based projects. Inspired by the Italian word for **"time"**, its name naturally complements `templ`, helping developers streamline component generation and manage CSS & JS workflows with ease.

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

- **[Plop.js](https://plopjs.com/)-inspired scaffolding** – Define templates and actions to quickly generate new components or variants.
- **Opinionated asset workflow** – A structured approach to managing CSS and JS within a `templ`-based project.
- **Configurable asset extraction** – Extracts CSS and JS from a designated source folder and injects them into `.templ` files in the corresponding destination folder, using guard markers.
- **Extensible template system** – Register custom template function providers (local or remote) to enhance Tempo's templating capabilities.
- **A fully Golang-native alternative to Node.js-based tools** – Achieves the same streamlined component and asset management without requiring Node.js.

## 💡 Motivation

While building a **UI component library in Golang with `templ`**, I faced two key challenges:

### The Problem

1. **Scaffolding new components** – Every component followed the same folder structure, but manually copying files and folders was inefficient. I initially used [Plop.js](https://plopjs.com/), but it required setting up a full Node.js project.
2. **Managing CSS & JS assets** – `templ` provides a great Go/HTML templating experience but lacks built-in tools for handling styles and scripts, making asset management cumbersome. (_See the ongoing discussion in [#740](https://github.com/a-h/templ/issues/740)._)

### The Solution

`tempo` solves both problems **natively in Go**, eliminating the need for Node.js while providing a **structured, opinionated workflow** for component and asset management.

#### ✨ Preserving the Developer Experience

With `tempo`, **CSS and JS files remain untouched during development**, allowing developers to continue using their preferred tools:

- **Linters & formatters** (e.g., Prettier, ESLint, Stylelint)
- **IDE features** like **syntax highlighting**, **error detection**, and **inline code suggestions** (e.g., in **VSCode**).
- **Existing workflows** remain intact—ensuring a familiar, efficient experience.

When you run `tempo`, **CSS and JS files are injected into `.templ` components automatically**, but the **original source files remain unchanged**. This approach **preserves developer productivity** while enabling seamless integration into templ-based projects.

## 💻 Installation

### go install

Requires Go v1.23+

```bash
go install github.com/indaco/tempo@latest
```

### Manually

Download the pre-compiled binaries from the [releases page](https://github.com/indaco/tempo/releases) and place it in a directory available in your system's _PATH_.

## 🚀 Usage

**1. Initialize tempo**

```bash
tempo init
```

Generates a `tempo.yaml` configuration file. Customize it to fit your project. See the [Configuration](#️-configuration) section for details.

**2. Define a Component or Variant**

```bash
tempo define component
tempo define variant
```

Generates templates for _components/variants_ inside `.tempo-files/templates/` along with an action JSON file inside `.tempo-files/actions/`.

**3. Create a Component**

```bash
tempo new component --name button
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
   tempo - A simple CLI for scaffolding components and managing assets in templ-based projects

USAGE:
   tempo [global options] [command [command options]]

VERSION:
   v0.1.0

COMMANDS:
   init, i            Initialize a Tempo project
   define, d          Define templates and actions for component or variant
   new, n             Generate a component or variant based on defined templates
   sync, s            Process & sync asset files into component templates
   register, r        Register is used to extend tempo
   clean              Remove temporary metadata generated by tempo run
   help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### init

Initialize a tempo project by generating a `tempo.yaml` configuration file. This file allows you to **customize how components and assets are generated and managed**.

For a full list of configuration options, See the [Configuration](#️-configuration) section for details.

### define

Define templates and actions for component or variant

```bash
tempo define component
tempo define variant
```

- **component** – Creates a scaffold for defining the structure of a new UI component, including a template and an action file.
- **variant** – Similar to component, but focused on defining a variant of an existing component (e.g., an _outline_ button variant).

These definitions are used later by `tempo new` to generate real instances of components.

<details>
<summary><strong>Flags</strong></summary>
<dl>
  <dt><code>--force</code></dt>
  <dd>Force overwriting if already exists (<em>default: false</em>)</dd>

  <dt><code>--dry-run</code></dt>
  <dd>Preview actions without making changes (<em>default: false</em>)</dd>

  <dt><code>--js</code></dt>
  <dd>Whether or not JS is needed for the component (<em>default: false</em>)</dd>
</dl>
</details>

### new

Generate a component or variant based on defined templates

```bash
tempo new component --name button
tempo new variant --name outline --component button
```

- **component** – Uses the templates and actions created with tempo define to generate a real component inside assets/ and components/.
- **variant** – Similar to component, but generates a new variant inside an existing component folder.

Example: Running `tempo new component --name dropdown` will generate:

- components/dropdown/ (with .templ and .go files).

<details>
<summary><strong>Component Flags</strong> (<code>tempo new component</code>)</summary>
<dl>
  <dt><code>--module</code> (<code>-m</code>) <em>value</em></dt>
  <dd>The name of the Go module being worked on</dd>

  <dt><code>--package</code> (<code>-p</code>) <em>value</em></dt>
  <dd>The Go package name where components will be generated (<em>default: components</em>)</dd>

  <dt><code>--assets</code> (<code>-a</code>) <em>value</em></dt>
  <dd>The directory where asset files (e.g., CSS, JS) will be generated (<em>default: assets</em>)</dd>

  <dt><code>--name</code> (<code>-n</code>) <em>value</em></dt>
  <dd>Name of the component</dd>

  <dt><code>--js</code></dt>
  <dd>Whether or not JS is needed for the component (<em>default: false</em>)</dd>

  <dt><code>--watermark</code></dt>
  <dd>Whether or not to include comments as a watermark in generated files (<em>default: false</em>)</dd>

  <dt><code>--force</code></dt>
  <dd>Force overwriting if the component already exists (<em>default: false</em>)</dd>

  <dt><code>--dry-run</code></dt>
  <dd>Preview actions without making changes (<em>default: false</em>)</dd>
</dl>
</details>

<details>
<summary><strong>Variant Flags</strong> (<code>tempo new variant</code>)</summary>

This command shares all the flags from `new component`, plus:

<dl>
  <dt><code>--component</code> (<code>-c</code>) <em>value</em></dt>
  <dd>Name of the component or entity to which this variant belongs</dd>
</dl>
</details>

### sync

Process & sync asset files into templ component files

```bash
tempo sync
```

This command scans the `input` folder for CSS and JS files and injects their content into the corresponding .templ files in the `output` folder:

- It extracts CSS and JS from files in the input folder.
- It injects them into `.templ` files, inside a section delimited by guard markers.
- This ensures components always have the latest CSS and JS without manual copying.

Whenever you update your CSS or JS files, simply run `tempo sync` to propagate the changes.

> [!TIP]
> **Live Reload:** If you're using a live reload tool like [air](https://github.com/air-verse/air), [templier](https://github.com/romshark/templier), or [watchexec](https://github.com/watchexec/watchexec), pass `--summary none` to reduce unnecessary output.

> [!NOTE]
> **Want to use tempo without scaffolding?**
> Check out [Using tempo sync as a Standalone Command](#-using-tempo-sync-as-a-standalone-command).

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

  <dt><code>--verbose-summary</code></dt>
  <dd>Show detailed information in the summary report (<em>default: false</em>)</dd>

  <dt><code>--report-file</code> (<code>--rf</code>) <em>value</em></dt>
  <dd>Export the summary to a JSON file</dd>
</dl>
</details>

### register

Register is used to extend tempo.

```bash
tempo register -h
tempo register -h functions
```

- **functions** – Register a function provider from a local go module path or a remote repository

```bash
tempo register functions --name sprig --url https://github.com/indaco/tempo-provider-sprig.git
```

<details>
<summary><strong>Functions Flags</strong> (<code>tempo register functions</code>)</summary>
<dl>
  <dt><code>--name</code> (<code>-n</code>) <em>value</em></dt>
  <dd>Name for the function provider</dd>

  <dt><code>--url</code> (<code>-u</code>) <em>value</em></dt>
  <dd>Repository URL</dd>

  <dt><code>--path</code> (<code>-p</code>) <em>value</em></dt>
  <dd>Path to a local go module provider</dd>
</dl>
</details>

### clean

Remove temporary metadata generated by `tempo sync`

```bash
tempo clean
```

- Deletes `.tempo-lastrun`, a file used to track asset processing timestamps.
- This command is useful if you need to reset the sync state before running `tempo sync` again.

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

- Default markers:

```templ
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */
```

- These can be customized in your `tempo.yaml` file under the `templates.guard_marker` setting.

> [!IMPORTANT]
> If no guard markers are present, `tempo sync` will **skip** the file, ensuring only intended sections are updated.

## ⚙️ Configuration

<details>
<summary><strong>📄 View Default <code>tempo.yaml</code> Configuration</strong></summary>

    # The root folder for Tempo files
    tempo_root: .tempo-files

    app:
      # The name of the Go module being worked on.
      # go_module:

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
      summary_format: long

    templates:
      # Whether or not to include comments as watermarks in generated files.
      watermark: true

      # A text placeholder or sentinel used in template files to mark auto-generated content.
      guard_marker: tempo

      # File extensions used for template files.
      extensions:
        - .gotxt
        - .gotmpl
        - .tpl

      # List of function providers for template processing.
      function_providers:
        # Example provider using a local path.
        # - name: default
        #   type: path
        #   value: ./providers/default
        #
        # Example provider from a Git repository.
        # - name: custom
        #   type: url
        #   value: https://github.com/user/custom-provider.git

</details>

### Configuration Options

#### Project Structure

| Key          | Default        | Description                                                 |
| :----------- | :------------- | :---------------------------------------------------------- |
| `tempo_root` | `.tempo-files` | Root folder where `tempo` stores its templates and actions. |

#### Application-Specific Settings

| Key              | Default      | Description                                              |
| :--------------- | :----------- | :------------------------------------------------------- |
| `app.go_module`  | _(empty)_    | Name of the Go module being worked on.                   |
| `app.go_package` | `components` | Go package name where components will be generated.      |
| `app.assets_dir` | `assets`     | Directory where asset files (CSS, JS) will be generated. |
| `app.with_js`    | `false`      | Whether JavaScript is required for the component.        |
| `app.css_layer`  | `components` | CSS layer name to associate with component styles.       |

#### Files Processing Options

| Key                        | Default | Description                                        |
| :------------------------- | :------ | :------------------------------------------------- |
| `processor.workers`        | `4`     | Number of concurrent workers.                      |
| `processor.summary_format` | `long`  | Summary format: `long`, `compact`, `json`, `none`. |

#### Templates Options

| Key                            | Default                     | Description                                                                                 |
| :----------------------------- | :-------------------------- | :------------------------------------------------------------------------------------------ |
| `templates.watermark`          | `true`                      | Whether to include comments as watermarks in generated files.                               |
| `templates.guard_marker`       | `tempo`                     | Placeholder used in templates to mark auto-generated content.                               |
| `templates.extensions`         | `.gotxt`, `.gotmpl`, `.tpl` | File extensions used for template files.                                                    |
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

📌 The **`tempo define component`** and **`tempo define variant`** commands generate:

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

| Variable        | Description                                                            |
| :-------------- | :--------------------------------------------------------------------- |
| `TemplatesDir`  | The root directory containing template files.                          |
| `ActionsDir`    | The root directory containing actions files.                           |
| `GoModule`      | The name of the Go module being worked on.                             |
| `GoPackage`     | The Go package name where components will be organized and generated.  |
| `ComponentName` | The name of the component being generated.                             |
| `VariantName`   | The name of the variant being generated (if applicable).               |
| `AssetsDir`     | The directory where asset files (CSS, JS) will be generated.           |
| `WithJs`        | Whether JavaScript is required for the component (true/false).         |
| `CssLayer`      | The CSS layer name associated with component styles.                   |
| `GuardMarker`   | Placeholder used in templates to mark auto-generated sections.         |
| `WatermarkTip`  | Whether to include watermark comments in generated files (true/false). |

#### 📌 Extending Template Functions

Tempo supports external function providers, allowing you to integrate additional helper functions into your templates.

See the full guide in [Extending Tempo](#-extending-tempo).

##### 📝 Example: Using Registered Functions in Templates

Example **component template file (component.templ.gotxt)**:

```go
package {{ .ComponentName | goPackageName }}

import (
    "{{ .GoModule }}/{{ .GoPackage | normalizePath | goPackageName }}/{{ .ComponentName | goPackageName }}/css"
)

templ {{ .ComponentName | goExportedName }}() {
    @css.{{ .ComponentName | goExportedName }}CSS()

    // continue here...
}
```

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

## 🔌 Extending Tempo

Tempo allows you to register external function providers, which means you can integrate additional helper functions into your templates.

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
