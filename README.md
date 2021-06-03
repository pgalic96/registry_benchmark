# CReB: Container Registry Benchmarking toolset

CReB is an extensible toolkit to run benchmarks against any container image registry (such as Docker Hub or AWS ECR) with both existing and generated workloads.

## Installation and Setup

Prerequisites:
- `go` (> 1.12)
- `python2.7`

1. Clone this repository and update your working directory.
2. Clone the CReB Trace Replayer repository somewhere on your system: `git clone git@github.com:pgalic96/docker-performance.git`
  a. Navigate to this repository and run `pip install -r requirements.txt`.
4. Configure the `config.yaml`. See `config-example.yaml` for the configurations used in our experiments.
5. [Optional] If you have access to the [DAS supercomputer](https://www.cs.vu.nl/das/) and want to use to run CReB in clustered mode, configure `das-config.yaml` as well.
6. [Optional] To run real-world traces download the IBM traces from our experiment artifacts on Zenodo: https://zenodo.org/record/4374309
7. Build the creb tool: `go build -o creb .` and add the resulting binary to your PATH.

## Usage

### push

Command `push` will generate a synthetic image and push it to the registries.

Example:
```bash
creb push
```

#### Notes for reviewers

In order to run this command you need to have the following fields set in the `config.yaml` : at least one `registry`, `pull-source-folder`, `image-generation` block.

### manpull

Command `manpull` will pull the manifest of the image. The image must have been pushed before running this command.

Example:
```bash
creb push
creb manpull
```

#### Notes for reviewers

In order to run this command you need to have the following fields set in the `config.yaml` : at least one `registry`, `pull-source-folder`, `image-generation` block. You also need to make sure the manifest was already pushed to the registry (which is done by setting `upload-manifest` flag in the `registry` block).

### layerpull

Command `layerpull` will pull the layers of an image. The image must have been pushed before running this command.

Example:
```bash
creb push
creb layerpull
```

#### Notes for reviewers

In order to run this command you need to have the following fields set in the `config.yaml` : at least one `registry`, `pull-source-folder`, `image-generation` block. You also need to make sure the layers were already pushed to the registry (via push command).

### trace-replayer

Command `trace-replayer` will configure and run the trace-replayer tool. To run CReB in the trace-replayer mode you will need the following:

1. The `docker-performance` repo present on your machine (see Installation);
2. The workload traces should be downloaded
3. The trace-replayer config in `config.yaml` needs to be set, pointing to the local docker-performance repo and the workload traces.

Example:
```bash
creb trace-replayer -d local
```

#### Notes for reviewers

In order to run the trace replayer locally you need to have the following configuration set:
In `config.yaml`: `trace-replayer` block, at least one `registry`.
