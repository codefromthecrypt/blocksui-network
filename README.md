# ⚠️ WIP: This project is being actively worked on.👷

Please see the [Contributing Guide](#-contributing) below.

<p align="center">
  <a href="https://blocksui.xyz/">
    <img alt="Blocks UI Protocol" src="assets/blocksui.png" width="320" />
  </a>
</p>

<h1 align="center">
  Blocks UI Protocol: Decentralized UI Software as an NFT
</h1>

<p align="center">
  🌎 💽 🤑
</p>

<p align="center">
  <strong>
    Decentralized. UI Software. As NFTs.
  </strong>
</p>

<p align="center">
  Providing an open and decentralized framework for building user interface software that is simple enough for anyone to use.
</p>

<p align="center">
  <a href="https://filecoin.io/" style="display: inline-block">
    <img src="assets/filecoin.svg" alt="Filecoin logo" width="40">
  </a>
  <a href="https://ipfs.tech/" style="display: inline-block">
    <img src="assets/ipfs.svg" alt="IPFS logo" width="120">
  </a>
  <a href="https://litprotocol.com/" style="display: inline-block">
    <img src="assets/lit-protocol.svg" alt="Lit Protocol logo" width="54">
  </a>
  <a href="https://polygon.technology/" style="display: inline-block">
    <img src="assets/polygon.svg" alt="Polygon logo" width="166">
  </a>
</p>

## Table of Contents

- [Contributing](#contributing)
  - [Types](#types)
  - [Branches](#branches)
  - [Commits](#commits)
  - [Pull Requests](#pull-requests)
  - [Merging Into Main](#merging-into-main)


## Blocks UI Protocol

- [Blocks UI App](https://github.com/crcls/blocksui-app)
- [Blocks UI Blocks](https://github.com/crcls/blocksui-blocks)
- [Blocks UI Contracts](https://github.com/crcls/blocksui-contract)
- [Blocks UI Network](https://github.com/crcls/blocksui-network)
- [Blocks UI SDK](https://github.com/crcls/blocksui-sdk)

## 🤝 Contributing

Code changes can fall into the types from the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.

### Types

Common types according to [commitlint-config-conventional (based on the Angular convention)](https://github.com/conventional-changelog/commitlint/tree/master/@commitlint/config-conventional#type-enum) can be:

- **build**—changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm).

- **chore**—other changes that don’t modify src or test files.

- **ci**—changes to our CI configuration files and scripts (example scopes: Travis, Circle, BrowserStack, SauceLabs).

- **docs**—documentation only changes.

- **feat**—a new feature.

- **fix**—a bug fix.

- **perf**—a code change that improves performance.

- **refactor**—a code change that neither fixes a bug nor adds a feature.

- **revert**—reverts a previous commit.

- **style**—changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc).

- **test**—adding missing tests or correcting existing tests.

### Branches

Branches are created from `main` and can follow the naming convention below. For common types, see [Types](#types).

Convention:

```shell
type/description
```

Example:

```shell
feat/add-xyz
```

### Commits

Conventional Commits are enforced using [commitlint](https://commitlint.js.org/) in a [husky](https://github.com/typicode/husky) pre-commit hook.

Convention:

```shell
type(scope?): description  #scope is optional; multiple scopes are supported (current delimiter options: "/", "\" and ",")
```

Example:

```shell
feat: add xyz
```

### Pull Requests

1.  **Title**

    Titles can follow the naming convention below and match the branch name. For common types, see [Types](#types).

    Convention:

    ```shell
    type: description
    ```

    Example:

    ```shell
    feat: add xyz
    ```

2.  **Body**

    When creating a new pull request, you will automatically see a template with a checklist in the body.

3.  **Reviewers**

    Add at least one reviewer.

4.  **Labels**

    Apply related labels.

### Merging Into Main

Always “Squash & merge” your commits into `main`.

## 🧐 License

Licensed under the [MIT License](./LICENSE).

## 💜 Thanks

Thanks go out to all of the many sponsors, [Filecoin](https://filecoin.io/), and [Moralis](https://moralis.io/).
