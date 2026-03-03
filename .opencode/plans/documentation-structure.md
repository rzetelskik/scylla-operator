# ScyllaDB Operator Documentation Structure

## Overview

Restructure documentation from API/resource-focused to user-journey-focused organization.

**Goals:**
- Follow user's natural deployment journey
- Add production readiness docs (issue #2916)
- Add ScyllaDB Manager docs (issue #1898)
- Apply Diataxis principles pragmatically

**Current state:** 86 markdown/rst files organized by API resources  
**New structure:** 10 user-journey sections with ~105 total files

---

## Table of Contents

### 1. Getting Started

```
getting-started/ - Getting Started
├── overview.md - ScyllaDB Operator Overview                                  [New] [Explanation]
└── deployment.md - Deploy ScyllaDB in Kubernetes                             [New] [Reference]
    # Summary of deployment steps with links to existing documents, not a step-by-step tutorial
```

### 2. Installation

```
installation/ - Installation
├── overview.md - Installation Overview                                       [Keep] [Explanation]
│
├── prerequisites/ - Prerequisites
│   ├── kubernetes-prerequisites.md - Kubernetes Prerequisites                [Enhance] [Reference]
│   │   # Generic Kubernetes prerequisites: Static CPU policy, node labels, packages
│   ├── platform-setup/ - Platform Setup
│   │   ├── setup-gke-single.md - Setup single GKE cluster                    [Extract from: quickstarts/gke.md] [How-to]
│   │   ├── setup-eks-single.md - Setup single EKS cluster                    [Extract from: quickstarts/eks.md] [How-to]
│   │   ├── setup-gke-multidc.md - Setup interconnected GKE clusters          [Move from: resources/common/multidc/gke.md] [How-to]
│   │   └── setup-eks-multidc.md - Setup interconnected EKS clusters          [Move from: resources/common/multidc/eks.md] [How-to]
│   └── multi-dc-prerequisites.md - Multi-DC Prerequisites                    [New] [Reference]
│       # Requirements for Multi-DC installation, e.g. ScyllaDB Manager in multi-dc setup
│
├── gitops.md - Install with kubectl (GitOps)                                 [Keep] [How-to]
├── helm.md - Install with Helm                                               [Keep] [How-to]
└── openshift.md - Install on Red Hat OpenShift                               [Keep] [How-to]
```

### 3. Configuration

```
configuration/ - Configuration
├── dedicated-node-pools.md - Dedicated Node Pools                            [New] [How-to] #2916
├── cpu-pinning.md - CPU Pinning                                              [New] [How-to] #2916
├── configure-sysctls.md - Kernel Parameters (sysctls)                        [Move from: management/sysctls.md] [How-to]
├── storage-configuration.md - Storage Configuration                          [New] [Explanation + How-to] #2916
│   # Includes configuring local SSDs using NodeConfig
├── resource-requirements-qos.md - Resource Requirements and QoS              [New] [Explanation + Reference] #2916
├── rlimit-nofile.md - Configure RLIMIT_NOFILE                                [New] [How-to] #2916
├── coredump-collection.md - Coredump Collection                              [New] [How-to] #2916
│
└── scylla-manager/ - ScyllaDB Manager
    ├── configure-agent-resources.md - Configure Manager Agent Resources      [New] [How-to + Reference] #1898
    └── configure-manager-registration.md - Configure Manager Registration    [New] [How-to] #1898
```

### 4. Deployment

```
deployment/ - Deployment
├── production-checklist.md - Production Deployment Checklist                 [New] [Reference] #2916
├── deploy-single-dc.md - Deploy Single-DC Cluster                            [Keep from: resources/scyllaclusters/basics.md] [Tutorial + How-to]
├── deploy-multi-dc.md - Deploy Multi-DC Cluster                              [Extract from: resources/scyllaclusters/multidc/multidc.md] [Explanation + How-to]
├── exposing.md - Expose ScyllaDB Clusters                                    [Move from: resources/common/exposing.md] [Explanation + How-to]
│
└── clients/ - Clients
    ├── discovery.md - Discovering ScyllaDB Nodes                             [Move from: resources/scyllaclusters/clients/discovering.md] [Explanation + How-to]
    ├── cql.md - Connect via CQL                                              [Move from: resources/scyllaclusters/clients/cql.md] [How-to]
    └── alternator.md - Connect via Alternator (DynamoDB)                     [Move from: resources/scyllaclusters/clients/alternator.md] [How-to]
```

### 5. Monitoring

```
monitoring/ - Monitoring
├── overview.md - ScyllaDB Monitoring Overview                                [Move from: management/monitoring/overview.md] [Explanation]
├── setup.md - Setup ScyllaDB Monitoring                                      [Move from: management/monitoring/setup.md] [How-to]
├── exposing-grafana.md - Expose Grafana                                      [Move from: management/monitoring/exposing-grafana.md] [How-to]
└── external-prometheus-on-openshift.md - Setup Monitoring on OpenShift       [Move from: management/monitoring/external-prometheus-on-openshift.md] [How-to]
```

### 6. Management

```
management/ - Management
├── replace-node.md - Replace a ScyllaDB Node                                 [Move from: resources/scyllaclusters/nodeoperations/replace-node.md] [How-to]
├── maintenance-mode.md - Maintenance Mode                                    [Move from: resources/scyllaclusters/nodeoperations/maintenance-mode.md] [How-to]
├── automatic-cleanup.md - Automatic Node Replacement                         [Move from: resources/scyllaclusters/nodeoperations/automatic-cleanup.md] [Explanation + How-to]
├── volume-expansion.md - Resize Cluster Storage                              [Move from: resources/scyllaclusters/nodeoperations/volume-expansion.md] [How-to]
├── data-cleanup.md - Automatic Data Cleanup                                  [Move from: management/data-cleanup.md] [Explanation + How-to]
├── bootstrap-sync.md - Synchronising bootstrap operations in ScyllaDB clusters [Move from: management/bootstrap-sync.md] [Explanation]
│
├── upgrading/ - Upgrading
│   ├── upgrade.md - Upgrade ScyllaDB Operator                                [Move from: management/upgrading/upgrade.md] [How-to]
│   └── upgrade-scylladb.md - Upgrade ScyllaDB                                [Move from: management/upgrading/upgrade-scylladb.md + resources/scyllaclusters/nodeoperations/scylla-upgrade.md] [How-to]
│
├── backup-restore/ - Backup and Restore
│   ├── schedule-backups.md - Schedule Backups                                [New] [How-to] #1898
│   ├── restore.md - Restore from Backup                                      [Move from: resources/scyllaclusters/nodeoperations/restore.md] [How-to]
│   └── manage-backup-tasks.md - Manage Backup Tasks                          [New] [How-to] #1898
│
├── repairs/ - Repairs
│   ├── schedule-repairs.md - Schedule Repairs                                [New] [How-to] #1898
│   └── manage-repair-tasks.md - Manage Repair Tasks                          [New] [How-to] #1898
│
└── networking/ - Networking
    └── ipv6/ - IPv6
        ├── tutorials/ - Tutorials
        │   └── ipv6-getting-started.md - Getting Started with IPv6           [Keep] [Tutorial]
        ├── concepts/ - Concepts
        │   └── ipv6-networking.md - IPv6 Networking Concepts                 [Keep] [Explanation]
        ├── how-to/ - How-To Guides
        │   ├── ipv6-configure.md - Configure Dual-Stack with IPv4            [Keep] [How-to]
        │   ├── ipv6-configure-ipv6-first.md - Configure Dual-Stack with IPv6 [Keep] [How-to]
        │   ├── ipv6-configure-ipv6-only.md - Configure IPv6-Only             [Keep] [How-to]
        │   ├── ipv6-migrate.md - Migrate Clusters to IPv6                    [Keep] [How-to]
        │   └── ipv6-troubleshoot.md - Troubleshoot IPv6 Issues               [Keep] [How-to]
        └── reference/ - Reference
            └── ipv6-configuration.md - IPv6 Configuration Reference          [Keep] [Reference]
```

### 7. Components

```
architecture/ - Components
├── overview.md - Operator Architecture                                       [Keep] [Explanation]
├── storage/ - Storage
│   ├── overview.md - Storage Architecture                                    [Keep] [Explanation]
│   └── local-csi-driver.md - Local CSI Driver                                [Keep] [Explanation + How-to]
├── tuning.md - Performance Tuning                                            [Keep] [Explanation]
└── manager.md - ScyllaDB Manager Integration                                 [Keep/Enhance] [Explanation]
```

### 8. Troubleshooting

```
troubleshooting/ - Troubleshooting
├── installation.md - Installation Issues                                     [Move from: support/troubleshooting/installation.md] [How-to]
├── known-issues.md - Known Issues                                            [Move from: support/known-issues.md] [Reference]
└── must-gather.md - Gather Diagnostic Data (must-gather)                     [Move from: support/must-gather.md] [How-to]
```

### 9. Support

```
support/ - Support
├── overview.md - Support Overview                                            [Keep] [Explanation]
├── support-matrix.md - Support Matrix                                        [Keep] [Reference]
└── releases.md - Releases                                                    [Move from: support/releases.md] [Reference]
```

### 10. Reference

```
reference/ - Reference
├── feature-gates.md - Feature Gates                                          [Keep] [Reference]
├── nodeconfig.md - NodeConfig Resource                                       [Move from: resources/nodeconfigs.md] [Reference]
├── scyllaoperatorconfig.md - ScyllaOperatorConfig Resource                   [Move from: resources/scyllaoperatorconfigs.md] [Explanation + Reference]
└── api/ - API Reference
    └── [Auto-generated API docs]                                             [Keep] [Reference]
        └── scylla.scylladb.com
            ├── NodeConfig
            ├── RemoteKubernetesCluster
            ├── RemoteOwner
            ├── ScyllaCluster
            ├── ScyllaDBCluster
            ├── ScyllaDBDataCenter
            ├── ScyllaDBDataCenterNodeStatusReport
            ├── ScyllaDBManagerClusterRegistration
            ├── ScyllaDBManagerTask
            ├── ScyllaDBMonitoring
            └── ScyllaOperatorConfig
```

### 11. _includes (Reusable Snippets)

```
_includes/ - Reusable Snippets
└── [Reusable content snippets]                                               [Rename from: .internal/]
```

---

## Documents to Drop

**After extraction:**
- `quickstarts/gke.md` → extract to installation/kubernetes-prerequisites.md
- `quickstarts/eks.md` → extract to installation/kubernetes-prerequisites.md

**Not recommended APIs (keep API reference only):**
- `resources/scylladbclusters/scylladbclusters.md`
- `resources/remotekubernetesclusters.md`

---

## New Documents Summary

**16 new documents total:**

### Production Readiness (7 docs - issue #2916)
1. `deployment/production-checklist.md` - Production deployment checklist
2. `configuration/dedicated-node-pools.md` - Creating dedicated node pools
3. `configuration/cpu-pinning.md` - Setting up CPU pinning
4. `configuration/storage-configuration.md` - Storage configuration (XFS, trimming)
5. `configuration/resource-requirements-qos.md` - Resource sizing and QoS classes
6. `configuration/rlimit-nofile.md` - Configuring RLIMIT_NOFILE
7. `configuration/coredump-collection.md` - Configuring coredump collection

### ScyllaDB Manager (7 docs - issue #1898)
8. `configuration/scylla-manager/configure-agent-resources.md` - Configuring Manager agent resources
9. `configuration/scylla-manager/configure-manager-registration.md` - Configuring Manager registration
10. `management/backup-restore/schedule-backups.md` - Scheduling backups with Manager
11. `management/backup-restore/manage-backup-tasks.md` - Managing backup tasks
12. `management/repairs/schedule-repairs.md` - Scheduling repairs with Manager
13. `management/repairs/manage-repair-tasks.md` - Managing repair tasks

### Getting Started (2 docs)
14. `getting-started/overview.md` - What is ScyllaDB Operator
15. `getting-started/deployment-guide.md` - Deployment phases and navigation

### Prerequisites (1 doc)
16. `installation/multi-dc-prerequisites.md` - Multi-DC prerequisites and Manager requirements

---

## Migration Map

| Source | Destination | Action |
|--------|-------------|--------|
| `resources/nodeconfigs.md` | `reference/nodeconfig.md` | Move |
| `management/sysctls.md` | `configuration/configure-sysctls.md` | Move |
| `resources/scyllaclusters/basics.md` | `deployment/deploy-single-dc.md` | Keep/Move |
| `resources/common/exposing.md` | `deployment/exposing.md` | Move |
| `resources/scyllaclusters/multidc/multidc.md` | `deployment/deploy-multi-dc.md` | Extract deployment parts |
| `resources/common/multidc/gke.md` | `installation/prerequisites/platform-setup/setup-gke-multidc.md` | Move |
| `resources/common/multidc/eks.md` | `installation/prerequisites/platform-setup/setup-eks-multidc.md` | Move |
| `resources/scyllaclusters/clients/discovering.md` | `deployment/clients/discovery.md` | Move |
| `resources/scyllaclusters/clients/cql.md` | `deployment/clients/cql.md` | Move |
| `resources/scyllaclusters/clients/alternator.md` | `deployment/clients/alternator.md` | Move |
| `management/monitoring/overview.md` | `monitoring/overview.md` | Move |
| `management/monitoring/setup.md` | `monitoring/setup.md` | Move |
| `management/monitoring/exposing-grafana.md` | `monitoring/exposing-grafana.md` | Move |
| `management/monitoring/external-prometheus-on-openshift.md` | `monitoring/external-prometheus-on-openshift.md` | Move |
| `resources/scyllaclusters/nodeoperations/replace-node.md` | `management/replace-node.md` | Move |
| `resources/scyllaclusters/nodeoperations/maintenance-mode.md` | `management/maintenance-mode.md` | Move |
| `resources/scyllaclusters/nodeoperations/automatic-cleanup.md` | `management/automatic-cleanup.md` | Move |
| `resources/scyllaclusters/nodeoperations/volume-expansion.md` | `management/volume-expansion.md` | Move |
| `resources/scyllaclusters/nodeoperations/restore.md` | `management/backup-restore/restore.md` | Move |
| `resources/scyllaclusters/nodeoperations/scylla-upgrade.md` | `management/upgrading/upgrade-scylladb.md` | Merge with existing |
| `management/data-cleanup.md` | `management/data-cleanup.md` | Keep (no path change) |
| `management/bootstrap-sync.md` | `management/bootstrap-sync.md` | Keep (no path change) |
| `management/upgrading/upgrade.md` | `management/upgrading/upgrade.md` | Keep (no path change) |
| `management/upgrading/upgrade-scylladb.md` | `management/upgrading/upgrade-scylladb.md` | Keep (no path change) |
| `resources/scyllaoperatorconfigs.md` | `reference/scyllaoperatorconfig.md` | Move |
| `support/must-gather.md` | `troubleshooting/must-gather.md` | Move |
| `support/troubleshooting/installation.md` | `troubleshooting/installation.md` | Move |
| `support/known-issues.md` | `troubleshooting/known-issues.md` | Move |
| `support/releases.md` | `support/releases.md` | Keep (no path change) |
| `quickstarts/gke.md` | `installation/prerequisites/platform-setup/setup-gke-single.md` | Extract platform setup |
| `quickstarts/eks.md` | `installation/prerequisites/platform-setup/setup-eks-single.md` | Extract platform setup |
| `installation/kubernetes-prerequisites.md` | `installation/prerequisites/kubernetes-prerequisites.md` | Move |
| `.internal/` | `_includes/` | Rename directory |

---

## References

- Issue #2916: https://github.com/scylladb/scylla-operator/issues/2916
- Issue #1898: https://github.com/scylladb/scylla-operator/issues/1898
- Diataxis Framework: https://diataxis.fr/
