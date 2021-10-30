# skbctf status

# configs

## Usage

- build: `make`
- prepare configs: refer to below
- `mkdir -p logs && sudo supervisorctl start all`

## checker

- Example config file of checker is [checker.example.conf.json](checker.example.conf.json)
- This file decides how test execution is done, such as execution interval, parallel execution, daemon mode, etc.
- You can also specify configuration by command-line option. The priority of options is `command-line > config file > default`.
- For usage of each options, run `./bin/main --help`.

## supervisord

- Exampe config file of supervisord is [supervisord.example.conf](supervisord.example.conf).
- This file decides how services run under supervisord.
- Main component is checker and badge-server.

# thanks

- This is heavily inspired by status-badge server of [TSGCTF2021](https://github.com/tsg-ut/tsgctf2021) (status-badge server itself is private). Thanks to its authors, especially [kcz146](https://twitter.com/kcz146).
- Badge-server depends on [shields.io](https://shields.io/).
