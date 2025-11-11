# DCN Netbox Infra Check

A Go application that compares VLANs between NAM and Netbox to identify
misconfigurations and discrepancies.

## Features

- Fetches VLAN and VxLAN data from NAM and Netbox APIs
- Detects missing or misconfigured VLANs in Netbox
- Finds VLANs with name mismatches between systems
- Identifies prefixes with incorrect infrastructure settings

## Configuration

The application reads configuration from files mounted in the container:

### ConfigMap (`/app/config/config.json`)

```json
{
    "netbox_url": "https://ipam.dcn.nhn.no",
    "nam_url": "https://dcn.nhn.no:3000",
    "checks": [
        {
            "netbox_site_id": 715,
            "infra": "prod",
            "dc_name": "nhn-trd2-vdc04"
        },
        {
            "netbox_site_id": 782,
            "infra": "prod",
            "dc_name": "nhn-bgo1-vdc15"
        }
    ]
}
```

### Secrets (mounted at `/app/secrets/`)

- `/secrets/netbox.secret` - Netbox API token
- `/secrets/nam.secret` - NAM API token
- `/secrets/esm.secret` - ESM password

## Local Development

### Prerequisites

- Go 1.25 or later
- Access to NAM, Netbox and ESM APIs

### Setup

1. Clone the repository:

```bash
cd /path/to/dcn-netbox-infra-check
```

2. Add your tokens:

```bash
echo "your-netbox-token" > secrets/netbox.secret
echo "your-nam-token" > secrets/nam.secret
echo "your-esm-password" > secrets/esm.secret
```

3. Edit the configuration:

```bash
nano config/config.json
```

### Building

Build the binary:

```bash
go build -o dcn-netbox-infra-check ./cmd/dcn-netbox-infra-check
```

## Docker

### Build Docker Image

```bash
docker build -t dcn-netbox-infra-check:latest .
```

## Kubernetes Deployment

### 1. Create Secrets

Update `deployments/kubernetes/secret.yaml` with your actual tokens:

```bash
# Option 1: Edit the file directly
kubectl apply -f deployments/kubernetes/secret.yaml

# Option 2: Create from command line
kubectl create secret generic dcn-netbox-infra-check-secrets \
  --from-literal=netbox-token="your-netbox-token" \
  --from-literal=nam-token="your-nam-token" \
  --from-literal=esm-password"your-esm-password"
```

### 2. Create ConfigMap

```bash
kubectl apply -f deployments/kubernetes/configmap.yaml
```

### 3. Deploy as CronJob (Scheduled)

The CronJob runs daily at 8:00 AM by default:

```bash
kubectl apply -f deployments/kubernetes/cronjob.yaml
```

To change the schedule, edit the `schedule` field in `cronjob.yaml`:

```yaml
spec:
    schedule: "0 8 * * *" # Cron format: minute hour day month weekday
```

### 4. Deploy as Manual Job

For one-time execution:

```bash
kubectl apply -f deployments/kubernetes/job.yaml
```

### View Logs

```bash
# For CronJob
kubectl logs -l app=dcn-netbox-infra-check --tail=100

# For specific job
kubectl logs job/dcn-netbox-infra-check-manual
```

### Delete Resources

```bash
kubectl delete cronjob dcn-netbox-infra-check
kubectl delete configmap dcn-netbox-infra-check-config
kubectl delete secret dcn-netbox-infra-check-secrets
```

## Configuration File Paths

The application expects configuration files at fixed paths:

- `/config/config.json` - Main configuration (URLs and check definitions)
- `/secrets/netbox.secret` - Netbox API token
- `/secrets/nam.secret` - NAM API token
- `/secrets/esm.secret` - ESM Password

These paths are designed to work with Kubernetes ConfigMaps and Secrets mounted
as volumes.

## Output

The application produces a detailed report for each DC check:

```
========================================
Processing DC: nhn-trd2-vdc04 (Site ID: 715, Infra: prod)
========================================

Fetching Netbox VLANs for site 715...
Fetching Netbox Prefixes for site 715...
Running checks...

===========================================================================
Vxlans i 'nhn-trd2-vdc04' som ikke er oppdatert i NAM etter flytting til nam-03 for 'prod'
===========================================================================
✗ [NAM VLAN ID 100] Netbox='vlan-nam-03' -> NAM='vlan-nam-01'

✓ Ingen avvik funnet!
```

If mismatches are found a request have been created in ESM

## Troubleshooting

### Common Issues

1. **Failed to load configuration**: Ensure config files are properly mounted in
   `/config` and `/secrets`
2. **Failed to fetch data**: Check API tokens and network connectivity
3. **No data returned**: Verify site IDs and DC names in configuration

### Debug Mode

To see more detailed output, check the pod logs:

```bash
kubectl logs -f <pod-name>
```
