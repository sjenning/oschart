# oschart
Bootchart for OpenShift Cluster Operators

![OSChart](https://raw.githubusercontent.com/sjenning/oschart/master/oschart.png)

## Building

```
go build ./cmd/oschart
```

## Running

Set your `KUBECONFIG` appropriately or using the `--kubeconfig` flag to pass it directly

```
./oschart
```

Then open up http://localhost:3000 to see the OSChart.

If it doesn't work, add `--logtostderr` to the flag to get more verbose logging.

## Data Collection and Offline Analysis

```
cd static
curl -OL http://localhost:3000/data.json
```

To view the data, launch a simple python webserver in the `static` directory
```
python3 -m http.server
```
and go to http://localhost:8000

