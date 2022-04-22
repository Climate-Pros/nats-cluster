# Global NATS Cluster with Security Enabled

[NATS](https://docs.nats.io/) is an open source messaging backend suited to many use cases and deployment scenarios. We use it for internal communications at Fly. This repo shows how to use it for your application.

This example creates a federated mesh of NATS servers that communicate over the private, encrypted IpV6 network available to all Fly organizations.

## Setup
This fork requires some additional setup to enable a system account. This is based on support from the NATS folks, fly.io folks, and my experimentation. Hopefully this can be improved in the future.

### Prerequisites

Install [nsc](https://docs.nats.io/using-nats/nats-tools/nsc). See the [nsc env helper README](./scripts/README.md) to set up your working directory as the store for credentials. Then set up a default environment. Choose your Operator name. 

```cmd
cp ./scripts/nsc-setup-env.sh /some/path
cd /some/path
chmod +x nsc-setup-env.sh
./nsc-setup-env.sh
source .nsc.env
nsc init -i
... complete nsc environment initialization ...
nsc generate config --nats-resolver --config-file nsc/config/server-jwt.conf
```

Set the following environment variables based on the data in nsc/config/server-jwt.conf:

NATS_OPERATOR_NAME from "Operator named ..."
NATS_OPERATOR_TOKEN from the "operator" property token
NATS_SYSTEM_ACCOUNT from the "system_account" property
NATS_RESOLVER_PAYLOAD in the "resolver_preload" property, you will see the "system_account" property as a key for a token. Use the token value only.

Once the app is deployed and Wireguard is set up (see below) edit the operator and push accounts up to your cluster:

```cmd
nsc edit operator --account-jwt-server-url "nats://{your fly app name}.internal:4222"
nsc push -A
```



1. `fly launch --no-deploy`

    > You'll be prompted for an app name. Hit return to let Fly generate an app name for you. Pick your target organization and a starting region.
    > Note: review fly.example.toml for configuration options to expose NATS 4222 to the internet with TLS enabled.

2. `flyctl deploy`

    > This will start NATS with a single node in your selected region.

3. Add more regions with `flyctl regions add <region>` or

    > For this demo, we set `ord`, `syd`, `cdg` regions.

```cmd
fly regions set ord syd cdg
```

4. Scale the application so it can place nodes in the regions.

```cmd
fly scale count 3
```

> Note: after scaling you will likely have to run `nsc push -A` again to update account JWTs on all available servers.

Then run `flyctl logs` and you'll see the virtual machines discover each other.

```
2020-11-17T17:31:07.664Z d1152f01 ord [info] [493] 2020/11/17 17:31:07.646272 [INF] [fdaa:0:1:a7b:abc:21de:af5f:2]:4248 - rid:1 - Route connection created
2020-11-17T17:31:07.713Z 21deaf5f cdg [info] [553] 2020/11/17 17:31:07.704807 [INF] [fdaa:0:1:a7b:81:d115:2f01:2]:34902 - rid:19 - Route connection created
2020-11-17T17:31:08.123Z 82fabc30 syd [info] [553] 2020/11/17 17:31:08.114852 [INF] [fdaa:0:1:a7b:81:d115:2f01:2]:4248 - rid:7 - Route connection created
2020-11-17T17:31:08.259Z d1152f01 ord [info] [493] 2020/11/17 17:31:08.241644 [INF] [fdaa:0:1:a7b:b92:82fa:bc30:2]:45684 - rid:2 - Route connection created
```

## Testing the cluster

Connect with [flyctl ssh](https://fly.io/docs/flyctl/ssh/) if needed, i.e. checking the generated `/etc/nats.conf` file. Note: if you haven't used `flyctl ssh` before, you will need to run `flyctl ssh establish` first to [set up your org root certificate](https://fly.io/docs/flyctl/ssh-establish/).

While the cluster is only accessible from inside the Fly network, you can use Fly's [Wireguard support](https://fly.io/docs/reference/private-networking/#step-by-step) to create a VPN into your Fly organisation and private network. Note: [Dealing With Defaults](https://fly.io/docs/reference/private-networking/#dealing-with-defaults) you should name your peer when creating, for example:

```cmd
flyctl wireguard create personal ord my-peer-name
```

With Wireguard enabled you can visit your monitoring endpoint

Then you can use tools such as [natscli](https://github.com/nats-io/natscli) to subscribe to topics, publish messages to topics and perform various tests on your NATS cluster. Install the tool first.

Once installed, create contexts that point at your NATS cluster:

http://appname.internal:8222



```cmd
nats context add fly.demo.sys --server appname.internal:4222 --description "My Cluster System Account" --creds nsc/nkeys/creds/{your created operator}/SYS/sys.creds
nats context add fly.demo --server appname.fly.dev:{port assigned in toml} --description "My Public Cluster" --creds nsc/nkeys/creds/{your created operator}/{your created account}/{your created account}.creds
nats context select
```

You can subscribe to a topic with `nats sub topicname`:

```cmd
nats sub fly.demo
```

And then, in another terminal sessions, we can use `nats pub topicname` to send either simple messages to that topic:

```cmd
nats pub fly.demo "Hello World"
```

Or send multiple messages:

```cmd
nats pub fly.demo "fly.demo says {{.Cnt}} @ {{.TimeStamp}}" --count=10
```

You're ready to start integrating NATS messaging into your other Fly applications.

## What to try next

1. [NATS streaming](https://docs.nats.io/nats-streaming-concepts/intro) offers persistence features, you can create a NATS streaming app by modifying this demo and adding volumes: `flyctl volume create`

2. Create a [NATS super cluster](https://docs.nats.io/nats-server/configuration/gateways) let you join multiple NATS clusters with gateways. If you want to run regional clusters, you can query the Fly DNS service to with `<region>.<app-name>.internal` to find server in specific regions.


## Discuss

You can discuss this example (and the paired 6pn-demo-chat example) on the [dedicated Fly Community topic](https://community.fly.io/t/new-examples-nats-cluster-and-6pn-demo-chat/562).

