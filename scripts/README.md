`nsc-setup-env.sh` generates `.nsc.env` as to set `nats` credentials environment variables to your current working directory.

Credit: https://github.com/wallyqs via Slack. (Thanks @wallyqs!)

```cmd
cp nsc-setup-env.sh /some/path/
cd /some/path
chmod +x ./nsc-setup-env.sh
./nsc-setup-env.sh
source .nsc.env
nsc env
+----------------------------------------------------------------------------------------------------------+
|                                             NSC Environment                                              |
+--------------------+-----+-------------------------------------------------------------------------------+
| Setting            | Set | Effective Value                                                               |
+--------------------+-----+-------------------------------------------------------------------------------+
| $NSC_CWD_ONLY      | No  | If set, default operator/account from cwd only                                |
| $NSC_NO_GIT_IGNORE | No  | If set, no .gitignore files written                                           |
| $NKEYS_PATH        | Yes | /some/path/nsc/nkeys                         |
| $NSC_HOME          | Yes | /some/path/nsc/accounts                      |
| $NATS_CA           | No  | If set, root CAs in the referenced file will be used for nats connections     |
|                    |     | If not set, will default to the system trust store                            |
| $NATS_KEY          | No  | If set, the tls key in the referenced file will be used for nats connections  |
| $NATS_CERT         | No  | If set, the tls cert in the referenced file will be used for nats connections |
+--------------------+-----+-------------------------------------------------------------------------------+
| From CWD           |     | No                                                                            |
| Default Stores Dir |     | /some/path/nsc/data/nats/nsc/stores          |
| Current Store Dir  |     | /some/path/nsc/data/nats/nsc/stores          |
| Current Operator   |     |                                                                               |
| Current Account    |     |                                                                               |
| Root CAs to trust  |     | Default: System Trust Store                                                   |
+--------------------+-----+-------------------------------------------------------------------------------+


```