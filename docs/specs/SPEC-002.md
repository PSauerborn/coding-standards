# SPEC-002: CI Pipeline

Spec ID: SPEC-002
Spec Date: 2026-08-01

## 1. Spec Statement

As a standards author I want a CI/CD pipeline that automates indexing and release of standards and build of the indexer so that I can create versioned releases

## 2. Context and Background

The current project has no CI pipelines. The indexer binary needs to be manually built and pushed, and standards need to be manually indexed. The goal of this spec is to create CI pipelines in Github that:

 - Build the indexer binary and commit it to the `bin/` directory
 - Index standards documents and commit the updated standards tree to the repository

Both pipelines should be manually triggered within Github, taking a semver as input from the user. Once the respective pipeline tasks have been ran and updated files have been committed, the repository should be tagged with the provided semver.

## 3. Scope Definitions

### 3.1 In Scope

- Create Github actions to automate build and tagging of indexer
- Create Github actions to automate indexing and creation of standards documents

### 3.2 Out of Scope

- Updating of any standards documents
- Updating the indexer source code

## 4. Requirements

- **REQ-1**: The repository must be configured with two separate Github actions: `release-indexer.yaml` and `release-standards.yaml`. Both Github actions must be manually triggered, and must require a `semver` argument as input arguments. See 6.2 for semver input formats.
 - **REQ-1.1**: Both pipelines must validate the semver provided before any other steps are ran. If a tag with the provided semver already exists, or it the semver is not a valid semver string, an error must be raised and the pipeline must be aborted.

- **REQ-2**: The `release-indexer.yaml` pipeline must build the indexer binary in the `indexer/` directory, place the binaries in the `bin/` folder, and commit the results before tagging the repository with the provided semver.
 - **REQ-2.1**: The `release-indexer.yaml` must build binaries for both amd64 and arm64 architectures on Linux, and arm64 for MacOS. arm64 for Mac OS executables must be built on macOS (`macos-latest` runners, which run on Apple Silicon) and amd64/arm64 executables for linux must be built on Linux (`ubuntu-latest` runners). macOS support is a requirement, not optional.
 - **REQ-2.2**: The executables for each OS/CPU architectures must be published as pipeline artifacts. The CPU architecture and OS must be included in the filename (see 6.1 for filename conventions).
 - **REQ-2.3**: Once all executables have been built, the file artifact for each must be exported from their respective jobs and placed into the `bin/` directory. The executables must be committed and pushed to the `main` branch using the `github-actions[bot]` user.
 - **REQ-2.4**: Once the executables have been pushed, the repository must be tagged with a `v{semver}-indexer` tag, where `{semver}` is the version provided as the input argument to the pipeline by the user.

- **REQ-3**: The `release-standards.yaml` pipeline must use the latest version of the indexer binary in the `bin/` directory to create the `standards-tree.yaml` index. The pipeline must run on a Linux AMD64 runner and use the `indexer-linux-amd64` binary.
 - **REQ-3.1**: The `release-standards.yaml` pipeline must run the `indexer` using `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` from the root level of the repository, creating the standards index.
 - **REQ-3.2**: The `standards-tree.yaml` file must be committed and pushed to the `main` branch using the `ci-bot` user.
 - **REQ-3.3**: Once the standards index has been pushed to the `main` branch, the repository must be tagged with a `v{semver}-standards` tag, where `{semver}` is the version provided as the input argument to the pipeline by the user.

## 5. Acceptance Criteria

- **AC-1** (REQ-2): Given that the tag `v0.2.0-indexer` does not exist
    When I run the `release-indexer.yaml` action with semver `0.2.0`
    Then a new version of the indexer binary should be built and committed to the main branch for arm64 and amd64 on Linux and arm64 for MacOS
    And the main branch should be tagged with `v0.2.0-indexer`

- **AC-2** (REQ-1.1): Given that the tag `v0.2.0-indexer` does exist
    When I run the `release-indexer.yaml` action with semver `0.2.0`
    Then the pipeline should fail
    And an error message should be displayed in the logs indicating that the version `0.2.0` already exists

- **AC-3** (REQ-1.1): Given that the tag `v0.2.0-indexer` does not exist
    When I run the `release-indexer.yaml` action with semver `0.2`
    Then the pipeline should fail
    And an error message should be displayed in the logs indicating that the version `0.2` is not a valid semver

- **AC-4** (REQ-3): Given that the tag `v0.2.0-standards` does not exist
    When I run the `release-standards.yaml` action with semver `0.2.0`
    Then a new version of the `standards-tree.yaml` file must be generated using the latest indexer in the `bin/` directory
    And the `standards-tree.yaml` file should be committed to the main branch
    And the main branch should be tagged with `v0.2.0-standards`

- **AC-5** (REQ-1.1): Given that the tag `v0.2.0-standards` does exist
    When I run the `release-standards.yaml` action with semver `0.2.0`
    Then the pipeline should fail
    And an error message should be displayed in the logs indicating that the version `0.2.0` already exists

- **AC-6** (REQ-1.1): Given that the tag `v0.2.0-standards` does not exist
    When I run the `release-standards.yaml` action with semver `0.2`
    Then the pipeline should fail
    And an error message should be displayed in the logs indicating that the version `0.2` is not a valid semver

 - **AC-7** (REQ-3): Given that the tag `v0.2.0-standards` does not exist
    When I run the `release-standards.yaml` action with semver `0.2.0`
    And the `standards-tree.yaml` file has not changed
    Then the pipeline should skip committing the updated index tree
    And the main branch should be tagged with `v0.2.0-standards`

## 6. Contracts and Constraints

### 6.1 Indexer Binary Filenames

Binary names for indexers must match the following format:

 - **arm64/macos**: `indexer-darwin-arm64`
 - **arm64/linux**: `indexer-linux-arm64`
 - **amd64**: `indexer-linux-amd64`

### 6.2 Input Semver

Input semvers must be of the format `1.2.3` i.e. `major.minor.patch` without any prefixes or suffixes.

### 6.3 GH Action Bot User

The bot user must be configured as follows

```txt
git config --global user.name "github-actions[bot]"
git config --global user.email "github-actions[bot]@users.noreply.github.com"
```

## 7. Edge Cases and Error Handling

 - **Unchanged index tree**: git commit on bytes-identical file will fail. If the `standards-tree.yaml` file has not changed, then the commit/push to `main` must be skipped. The repository must still be tagged with the new semver tag.

## 8. Infrastructure Requirements

N/A

## 9. External Resources

N/A
