# About jao

jao is a self-hosted app that installs and manages VPN protocols on a server. It lives on the server, runs the VPN protocols in Docker containers, and lets other people create their own connection keys.

## The problem

Each self-hosted VPN solution (Outline, AmneziaWG, XRay/VLESS, WireGuard, and others) comes with its own management tool and its own way to give keys to users. If you run more than one protocol, or more than one server, you must use many tools and do many manual steps.

Key delivery is the worst part. Each time a friend asks for a VPN, the server owner must create a config by hand and send it to the friend. If the friend needs a new key later, the owner must do it again.

## The idea

jao puts the protocols and the key management in one app on the server. The server owner sets up the server once. After that, friends create keys themselves, and the owner does not need to take part.

The main scenario today:

1. A friend asks the owner for a VPN.
2. The owner gets a new VPS and installs jao on it.
3. The owner uses jao to install one or more VPN containers on the VPS.
4. The owner gives the friend a link to their account on this node.
5. The friend opens the link and creates new keys when they need them.

This scenario is the minimum that jao must do.

## Concepts

- **Node**: one server with jao installed on it.
- **Owner**: the person who installs jao and controls the node.
- **User**: a person who gets access to a node from the owner, for example a friend.
- **Account link**: the link that the owner gives to a user. It opens the user's account on the node.
- **Protocol**: a VPN or proxy technology, for example AmneziaWG or Outline.
- **Protocol container**: a Docker container on the node that runs one protocol.
- **Key**: a connection config that a user imports into a VPN client app.

## Principles

- **A node is self-contained.** Each node has its own data and works alone. A node does not need any other node or central service to operate.
- **Decentralization comes first.** jao has no shared database and no required central server. When nodes connect to each other, this connection adds features but does not create a dependency.
- **Setup is short.** It must be possible to install jao on a new VPS and give a friend access in a few minutes.
- **Protocols are pluggable.** Support for a new protocol must not need changes in the core of the app.
- **Users stay anonymous.** jao does not need personal data about users to give them access.

## Future direction

When there are many nodes and many friends, the minimal scenario is not sufficient. We want to add these features later:

- **Access across nodes**: the owner controls which users have access to which nodes.
- **Reassignment**: the owner moves a user from one node to another, for example when a server is blocked or shut down.
- **A single control plane**: one place where the owner sees all nodes, their metrics (for example CPU load), and their protocol containers.
- **Remote key provisioning**: the owner creates keys for a user on any node from the control plane.

The control plane must obey the decentralization principle. It is a view and a control tool over independent nodes. If the control plane is unavailable, each node and its users continue to work.

## Out of scope

- jao does not create VPS instances. The owner gets the server from a hosting provider.
- jao does not manage servers that do not run jao.
- jao does not do billing or payments.
