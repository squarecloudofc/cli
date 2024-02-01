Square Cloud CLI.
This package provides a direct way to interact with the official Square Cloud API.

### Installation
To install the CLI, just run the following command in your terminal:

macOS, Linux, and WSL:
```bash
curl -fsSL https://cli.squarecloud.app/install | sh
```

Windows | need [npm](https://www.npmjs.com/) installed:
```bash
npm install -g @squarecloud/cli
```

Or access the @squarecloud/cli [npm page](https://www.npmjs.com/package/@squarecloud/cli) for more information.

### Commands 
List of all commands available in the CLI:

| Command | Description                                                           |
| ------- | --------------------------------------------------------------------- |
| app     | Perform actions with your applications                                |
| apps    | List all your Square Cloud applications                               |
| commit  | Commit your application to Square Cloud                               |
| help    | Get help about any command                                            |
| login   | Log in to Square Cloud                                                |
| whoami  | Print user information associated with the current Square Cloud Token |
| zip     | Zip the current folder                                                |

### Inside the app command
List of all commands available in the app command:

| Command | Description                         |
| ------- | ----------------------------------- |
| delete  | Delete your application             |
| restart | Restart your application            |
| start   | Start your application              |
| status  | Show the status of your application |
| stop    | Stop your application               |

### Update
To update the CLI, just run the following command in your terminal:

macOS, Linux, and WSL:
```bash
curl -fsSL https://cli.squarecloud.app/install.sh | sh
```

Windows | need [npm](https://www.npmjs.com/) installed:
```bash
squarecloud update
```