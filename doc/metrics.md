## Prometheus Metrics exposed by upgrade-manager
upgrade-manager exposes different Prometheus metrics that can be used for alerting/monitoring.
To scrape these metrics, you can refer to the default ServiceMonitor deployed with the chart.

- `upgrade_manager_software_obsolescence_score` (gaugeVec): obsolescence score for softwares discovered by upgrade-manager 
- `upgrade_manager_software_process_error` (counterVec): count of errors while processing softwares
- `upgrade_manager_total_software_found` (gauge): total number of softwares found in the auto-discovery process
- `upgrade_manager_total_software_load_success` (gauge): total count of softwares with successfully loaded version candidates
- `upgrade_manager_total_software_obsolescence_score_compute_success` (gauge): total count of softwares with successfully computed obsolescence scores.
- `upgrade_manager_main_loop_execution_time` (gauge): Time taken by the last main loop execution (to find softwares, find new versions and compute scores)"
