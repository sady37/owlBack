/Users/sady3721/project/owlBack
├── AUTH_STREAM_IMPLEMENTATION_SUMMARY.md
├── check_data_flow.sh
├── check_pin_hash.sql
├── check_ports.sh
├── DATABASE_EXPORT_GUIDE.md
├── DATABASE_STRUCTURE_EXPLANATION.md
├── DELIVERY_CHECKLIST.md
├── deploy-config.sh
├── deploy-quick-start.md
├── deploy.sh
├── DEPLOYMENT.md
├── docker-compose.logging.yml
├── docker-compose.yml
├── docs
│   ├── 01_Development_Progress.md
│   ├── 03_Development_Plan_Updated.md
│   ├── 07_Sleepace_Data_Source_Clarification.md
│   ├── 09_Sleepace_v1.0_Architecture_Analysis.md
│   ├── 10_Sleepace_Data_Flow_v1.5.md
│   ├── 11_Sleepace_Unified_Data_Flow_Implementation.md
│   ├── 12_Sensor_Fusion_Implementation.md
│   ├── 13_Alarm_Fusion_Implementation.md
│   ├── alarm_rule.md
│   ├── ARCHITECTURE_DESIGN.md
│   ├── ARCHITECTURE_IMPLEMENTATION_COMPARISON.md
│   ├── BOTTOM_UP_DESIGN.md
│   ├── BRANCH_ID_FILTER_STATISTICS.md
│   ├── CARD_UPDATE_ARCHITECTURE_OPTIONS.md
│   ├── CAREGIVER_CARD_ACCESS_FLOW.md
│   ├── design
│   │   ├── Current_Branch_Design.md
│   │   └── User_AccountSettings_API_Design.md
│   ├── DEVICE_BUSINESS_ACCESS_AND_MONITORING_ENABLED_LOGIC.md
│   ├── handler_analysis
│   │   ├── HANDLER_ANALYSIS_ALARM_CLOUD_SERVICE.md
│   │   ├── HANDLER_ANALYSIS_AUTH_SERVICE.md
│   │   ├── HANDLER_ANALYSIS_DEVICE_SERVICE.md
│   │   ├── HANDLER_ANALYSIS_ROLE_PERMISSION_SERVICE.md
│   │   ├── HANDLER_ANALYSIS_ROLE_SERVICE.md
│   │   ├── HANDLER_ANALYSIS_TAG_SERVICE.md
│   │   ├── HANDLER_ANALYSIS_USER_SERVICE.md
│   │   └── HANDLER_REFACTORING_ANALYSIS_TEMPLATE.md
│   ├── IMPLEMENTATION_SUMMARY.md
│   ├── internal
│   │   ├── repository
│   │   │   └── ALARM_EVENTS_REPOSITORY_UNIFICATION.md
│   │   └── service
│   │       ├── ALARM_EVENT_SERVICE_ANALYSIS.md
│   │       ├── ALARM_EVENT_SERVICE_INTERFACE_DESIGN.md
│   │       ├── ALARM_EVENT_TYPE_CONVERSION.md
│   │       ├── RESIDENT_SERVICE_DESIGN.md
│   │       ├── RESIDENT_SERVICE_IMPLEMENTATION_STATUS.md
│   │       ├── UNIT_SERVICE_ANALYSIS.md
│   │       ├── UNIT_SERVICE_COMPLETE.md
│   │       ├── UNIT_SERVICE_IMPLEMENTATION.md
│   │       ├── UNIT_SERVICE_VALIDATION.md
│   │       ├── USER_SERVICE_ANALYSIS.md
│   │       ├── USER_SERVICE_INTERFACE_DESIGN.md
│   │       └── USER_SERVICE_VALIDATION.md
│   ├── IOT_TIMESERIES_TABLE_STRUCTURE_OPTIMIZATION.md
│   ├── log.md
│   ├── RADAR_FALL_ALARM_FLOW.md
│   ├── Radar_HTTPS_MQTT_Protocol_Formatted.md
│   ├── repository_design
│   ├── reviews
│   │   ├── CHATGPT_ROUND1_PROMPT_COMPLETE.md
│   │   ├── CHATGPT_ROUND1_PROMPT.md
│   │   ├── chatgpt_round1_sensor_fusion.md
│   │   ├── README.md
│   │   ├── ROUND1_READY.md
│   │   ├── SUBMIT_TO_CHATGPT_NOW.md
│   │   └── SUBMIT_TO_CLAUDE.md
│   ├── role.md
│   ├── service_design
│   ├── system_architecture_complete.md
│   └── test
│       ├── User_AccountSettings_Test_Guide.md
│       └── User_AccountSettings_Test_Results.md
├── examples
│   └── auth_encoder_example.go
├── exports
│   └── owlrd_backup_20260109_003115.sql.gz
├── FILE_INDEX.md
├── FINAL_DELIVERY_REPORT.md
├── find_test_ids.sql
├── GetResidentAccountSettings_flow_analysis.md
├── logging
│   ├── grafana-datasources.yml
│   └── promtail-config.yml
├── mqtt
│   ├── config
│   │   └── mosquitto.conf
│   ├── data
│   │   └── mosquitto.db
│   └── log
│       └── mosquitto.log
├── owl-common
│   ├── card
│   │   ├── creator.go
│   │   ├── repository.go
│   │   ├── types.go
│   │   └── utils.go
│   ├── config
│   │   └── config.go
│   ├── database
│   │   └── postgres.go
│   ├── encode
│   │   ├── auth_encoder_test.go
│   │   ├── auth_encoder.go
│   │   ├── AUTH_STREAM_INTEGRATION_GUIDE.md
│   │   ├── common.go
│   │   ├── config
│   │   │   ├── 06_FHIR_Simple_Conversion_Guide.md
│   │   │   ├── Radar_HTTPS_MQTT_Protocol_Formatted.md
│   │   │   ├── RADAR_SNOMED_CATEGORY_MAPPING.md
│   │   │   ├── README_RADAR_CONVERT.md
│   │   │   └── sleepace_convert_table.json
│   │   ├── MAPPING_TABLE_COMPLETE.md
│   │   ├── radar_convert.go
│   │   ├── RADAR_DECODER_ANALYSIS.md
│   │   ├── radar_encoder.go
│   │   ├── RADAR_REDIS_STREAM_FORMAT_STANDARD.md
│   │   ├── RADAR_REDIS_STREAM_FORMAT.md
│   │   ├── README.md
│   │   ├── sleepace_convert.go
│   │   ├── SLEEPACE_ENCODER_IMPLEMENTATION.md
│   │   ├── sleepace_encoder.go
│   │   └── snomed_mapping.go
│   ├── errors
│   ├── go.mod
│   ├── go.sum
│   ├── logger
│   │   └── logger.go
│   ├── mqtt
│   │   └── client.go
│   └── redis
│       ├── client.go
│       └── streams.go
├── owlBack.code-workspace
├── PORT_CONFIG_CHECK_SUMMARY.md
├── PORTS_QUICK_REFERENCE.md
├── query_admin_pin_hash.sql
├── QUICK_REFERENCE.md
├── README_REDIS_AUTH_STREAM.md
├── README_SERVICES.md
├── REDIS_AUTH_STREAM_IMPLEMENTATION.md
├── REMOTE_DEPLOYMENT_COMPLETE.md
├── remote-setup.sh
├── scripts
│   ├── analyze_db_files.sh
│   ├── analyze_local_migrate.sh
│   ├── check
│   │   ├── go.mod
│   │   └── go.sum
│   ├── check_local_db_schema.sh
│   ├── check_migrate_status.sh
│   ├── check_remaining_migrates.sh
│   ├── check_sql_syntax_issues.sh
│   ├── clean_local_migrate.sh
│   ├── cleanup_db_files.sh
│   ├── cleanup_migrate_files.sh
│   ├── export_database.sh
│   ├── import_database_on_remote.sh
│   ├── independent-verify.sh
│   ├── init-db.sh
│   ├── sync_database_to_remote.sh
│   ├── test_user_account_settings_simple.sh
│   ├── test_user_account_settings.sh
│   ├── verify_migrate_merged.sh
│   └── verify.sh
├── start_owlback.sh
├── stop_owlback.sh
├── test_account_settings.sh
├── test_doctor.sh
├── test_radar_data.sh
├── test_results_final.log
├── test_results_full.log
├── test_results_with_login.log
├── test_results.log
├── tree-owlback.md
├── verify_account_settings.sql
├── wisefido-ai
│   ├── ALARM_EVENT_WRITE.md
│   ├── BACKEND_COMPARISON.md
│   ├── BED_REPOSITORY_ANALYSIS.md
│   ├── cmd
│   │   └── wisefido-ai
│   │       └── main.go
│   ├── DEVICE_SERVICE_DESIGN.md
│   ├── docs
│   │   ├── EVENT1_EVALUATION_FILTERS.md
│   │   ├── EVENT3_BATHROOM_FALL_ANALYSIS.md
│   │   ├── EVENT3_BATHROOM_FALL_DESIGN.md
│   │   └── RADAR_EVENT_INTEGRATION.md
│   ├── FILE_STRUCTURE_ANALYSIS.md
│   ├── go.mod
│   ├── go.sum
│   ├── IMPLEMENTATION_PLAN.md
│   ├── IMPLEMENTATION_SUMMARY.md
│   ├── internal
│   │   ├── config
│   │   │   ├── config_test.go
│   │   │   └── config.go
│   │   ├── consumer
│   │   │   ├── cache_consumer.go
│   │   │   ├── cache_manager_test.go
│   │   │   ├── cache_manager.go
│   │   │   ├── event_consumer.go
│   │   │   └── state_manager.go
│   │   ├── evaluator
│   │   │   ├── alarm_event_builder_test.go
│   │   │   ├── alarm_event_builder.go
│   │   │   ├── evaluator.go
│   │   │   ├── event1_bed_fall.go
│   │   │   ├── event1_helpers.go
│   │   │   ├── event2_sleepad_reliability.go
│   │   │   ├── event3_bathroom_fall.go
│   │   │   ├── event3_helpers.go
│   │   │   └── event4_sudden_disappear.go
│   │   ├── models
│   │   │   ├── alarm_config.go
│   │   │   ├── alarm_event.go
│   │   │   ├── iot_data_message.go
│   │   │   └── realtime_data.go
│   │   ├── repository
│   │   │   ├── alarm_cloud.go
│   │   │   ├── alarm_device.go
│   │   │   ├── alarm_events_test.go
│   │   │   ├── alarm_events.go
│   │   │   ├── card_test.go
│   │   │   ├── card.go
│   │   │   ├── config_version.go
│   │   │   ├── device.go
│   │   │   ├── iot_timeseries.go
│   │   │   └── room.go
│   │   └── service
│   │       ├── alarm_event_service.go
│   │       └── alarm.go
│   ├── ISSUES_ANALYSIS.md
│   ├── main
│   ├── QUICK_START.md
│   ├── RADAR_SLEEPACE_SERVICE_ANALYSIS.md
│   ├── README.md
│   ├── REPOSITORY_LAYER_SUMMARY.md
│   ├── REQUIREMENTS_ANALYSIS.md
│   ├── RUN_TEST.md
│   ├── scripts
│   │   ├── run_test.sh
│   │   ├── run_tests.sh
│   │   ├── test_alarm_events_repo.sh
│   │   ├── test_run.sh
│   │   └── verify_setup.sh
│   ├── SERVICE_DECISION_MATRIX.md
│   ├── SERVICE_DESIGN_PATTERNS.md
│   ├── SERVICE_DESIGN_PLAN.md
│   ├── SERVICE_LAYER_COMPLETE_DESIGN.md
│   ├── SERVICE_LAYER_SYSTEMATIC_ANALYSIS.md
│   ├── SERVICE_REQUIREMENTS_SUMMARY.md
│   ├── start-test.sh
│   ├── TEST_SUMMARY.md
│   ├── TESTING_GUIDE.md
│   ├── TROUBLESHOOTING.md
│   ├── UNIT_TEST_GUIDE.md
│   ├── VERIFY.md
│   ├── VITALFOCUS_SERVICE_DESIGN.md
│   ├── wisefido-ai
│   └── wisefido-alarm
├── wisefido-alarm
│   ├── internal
│   │   ├── fusion
│   │   └── inspection
│   └── pkg
├── wisefido-card-aggregator
│   ├── cmd
│   │   └── wisefido-card-aggregator
│   │       └── main.go
│   ├── coverage.out
│   ├── DATA_AGGREGATION_SUMMARY.md
│   ├── diagnose_card_generation.sh
│   ├── docs
│   │   ├── CARD_DESIGN_ANALYSIS.md
│   │   ├── CARD_FIELD_GENERATION_LOGIC.md
│   │   ├── CARD_NAME_STATUS_INDICATORS_DISCUSSION.md
│   │   ├── CARD_POLLING_OPTIMIZATION_DISCUSSION.md
│   │   ├── CARD_UPDATE_STRATEGIES.md
│   │   ├── CURRENT_CARD_NAME_LOGIC_ANALYSIS.md
│   │   ├── DATA_AGGREGATION_IMPLEMENTATION.md
│   │   ├── EVENT_DRIVEN_IMPLEMENTATION.md
│   │   ├── EVENT_TRIGGER_MECHANISM.md
│   │   ├── FACILITY_BED_UNBOUND_CHECK_DISCUSSION.md
│   │   ├── FACILITY_UNOCCUPIED_LOGIC.md
│   │   ├── IMPLEMENTATION_SUMMARY.md
│   │   ├── MONITORING_DISABLED_CARD_NAME_DISCUSSION.md
│   │   ├── PARALLEL_EVENT_POLLING_ANALYSIS.md
│   │   ├── SIMPLIFIED_UPDATE_STRATEGY_DISCUSSION.md
│   │   ├── STARTUP_CARD_CREATION.md
│   │   └── TIMEZONE_DECISION.md
│   ├── go.mod
│   ├── go.sum
│   ├── IMPLEMENTATION_CHECKLIST.md
│   ├── internal
│   │   ├── aggregator
│   │   │   ├── cache_manager_test.go
│   │   │   ├── cache_manager.go
│   │   │   ├── card_creator_test.go
│   │   │   ├── data_aggregator_test.go
│   │   │   ├── data_aggregator.go
│   │   │   ├── kv_fake_test.go
│   │   │   └── kv.go
│   │   ├── alarm
│   │   │   └── alarm_handler.go
│   │   ├── config
│   │   │   ├── config_test.go
│   │   │   └── config.go
│   │   ├── consumer
│   │   │   ├── config_consumer.go
│   │   │   ├── event_consumer.go
│   │   │   └── iot_stream_consumer.go
│   │   ├── fusion
│   │   │   └── sensor_fusion.go
│   │   ├── models
│   │   │   ├── alarm_event.go
│   │   │   ├── iot_data.go
│   │   │   ├── iot_timeseries.go
│   │   │   └── vital_focus_card.go
│   │   ├── repository
│   │   │   ├── alarm_device.go
│   │   │   ├── alarm_events.go
│   │   │   ├── card_device.go
│   │   │   ├── card_info.go
│   │   │   ├── card_test.go
│   │   │   ├── card.go
│   │   │   ├── iot_timeseries.go
│   │   │   ├── routing_test.go
│   │   │   └── routing.go
│   │   └── service
│   │       ├── aggregator_test.go
│   │       └── aggregator.go
│   ├── main
│   ├── pkg
│   ├── README.md
│   ├── start_service.sh
│   └── wisefido-card-aggregator
├── wisefido-card-manage
│   ├── cmd
│   │   └── wisefido-card-manage
│   │       └── main.go
│   ├── crontab.example
│   ├── FUNCTIONAL_ANALYSIS.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── config
│   │   │   └── config.go
│   │   ├── http
│   │   │   ├── handler.go
│   │   │   └── router.go
│   │   ├── repository
│   │   │   ├── card_info.go
│   │   │   ├── card_test.go
│   │   │   ├── card.go
│   │   │   ├── routing_test.go
│   │   │   └── routing.go
│   │   └── service
│   │       └── card_service.go
│   ├── REMOVAL_SUMMARY.md
│   └── scripts
│       └── create-all-cards.sh
├── wisefido-data
│   ├── ARCHITECTURE_LAYER_PRINCIPLES.md
│   ├── CAREGIVERS_BUSINESS_FLOW.md
│   ├── cmd
│   │   ├── apply-migration
│   │   │   └── main.go
│   │   ├── check-branch-only
│   │   │   └── main.go
│   │   ├── check-residents
│   │   │   └── main.go
│   │   ├── check-role-permissions
│   │   │   └── main.go
│   │   ├── check-role-permissions-schema
│   │   │   └── main.go
│   │   ├── execute-sql
│   │   │   └── main.go
│   │   ├── test-permission-utils
│   │   │   └── main.go
│   │   ├── verify-migration
│   │   │   └── main.go
│   │   └── wisefido-data
│   │       └── main.go
│   ├── CODE_REVIEW_REPORT.md
│   ├── Dockerfile
│   ├── docs
│   │   ├── vital_focus_service_design_v2.md
│   │   └── vital_focus_service_design_v3.md
│   ├── 启用Doctor.md
│   ├── go.mod
│   ├── go.sum
│   ├── http.test
│   ├── internal
│   │   ├── config
│   │   │   └── config.go
│   │   ├── domain
│   │   │   ├── alarm_cloud_update.go
│   │   │   ├── alarm_cloud.go
│   │   │   ├── alarm_device_update.go
│   │   │   ├── alarm_device.go
│   │   │   ├── alarm_event_update.go
│   │   │   ├── alarm_event.go
│   │   │   ├── bed_update.go
│   │   │   ├── bed.go
│   │   │   ├── branch_update.go
│   │   │   ├── branch.go
│   │   │   ├── building_update.go
│   │   │   ├── building.go
│   │   │   ├── card.go
│   │   │   ├── config_version_update.go
│   │   │   ├── config_version.go
│   │   │   ├── device_store_update.go
│   │   │   ├── device_store.go
│   │   │   ├── device_update.go
│   │   │   ├── device.go
│   │   │   ├── iot_timeseries.go
│   │   │   ├── resident_caregiver_update.go
│   │   │   ├── resident_caregiver.go
│   │   │   ├── resident_contact_update.go
│   │   │   ├── resident_contact.go
│   │   │   ├── resident_phi_update.go
│   │   │   ├── resident_phi.go
│   │   │   ├── resident_update.go
│   │   │   ├── resident.go
│   │   │   ├── role_permission.go
│   │   │   ├── role.go
│   │   │   ├── room.go
│   │   │   ├── round_detail.go
│   │   │   ├── round.go
│   │   │   ├── sleepace_report.go
│   │   │   ├── snomed_mapping.go
│   │   │   ├── tenant.go
│   │   │   ├── unit.go
│   │   │   ├── update_field.go
│   │   │   ├── user_branch.go
│   │   │   └── user.go
│   │   ├── handler
│   │   ├── http
│   │   │   ├── admin_alarm_cloud_handler.go
│   │   │   ├── admin_alarm_handlers.go
│   │   │   ├── admin_device_store_impl.go
│   │   │   ├── admin_other_handlers.go
│   │   │   ├── admin_role_permissions_handler.go
│   │   │   ├── admin_roles_handler.go
│   │   │   ├── admin_roles_handlers.go
│   │   │   ├── admin_tenants_handlers.go
│   │   │   ├── admin_units_devices_handlers.go
│   │   │   ├── admin_units_devices_impl.go
│   │   │   ├── alarm_event_handler.go
│   │   │   ├── auth_handler_test.go
│   │   │   ├── auth_handler.go
│   │   │   ├── auth_store.go
│   │   │   ├── branches_handler.go
│   │   │   ├── card_overview_e2e_test.go
│   │   │   ├── card_overview_handler_test.go
│   │   │   ├── card_overview_handler.go
│   │   │   ├── device_handler.go
│   │   │   ├── device_monitor_settings_handler.go
│   │   │   ├── device_store_excel.go
│   │   │   ├── device_store_handler.go
│   │   │   ├── doctor_handler.go
│   │   │   ├── permission_utils.go
│   │   │   ├── radar_handler.go
│   │   │   ├── resident_handler.go
│   │   │   ├── result.go
│   │   │   ├── router.go
│   │   │   ├── sleepace_report_handler_test.go
│   │   │   ├── sleepace_report_handler.go
│   │   │   ├── stub_handler_base.go
│   │   │   ├── stub_handlers.go
│   │   │   ├── unit_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── util.go
│   │   │   ├── vital_focus_handlers_test.go
│   │   │   └── vital_focus_handlers.go
│   │   ├── middleware
│   │   ├── models
│   │   │   ├── pagination.go
│   │   │   └── vital_focus.go
│   │   ├── mqtt
│   │   │   └── sleepace_broker.go
│   │   ├── notifier
│   │   │   └── config_notifier.go
│   │   ├── repository
│   │   │   ├── alarm_cloud_repo.go
│   │   │   ├── alarm_device_repo.go
│   │   │   ├── alarm_events_repo.go
│   │   │   ├── auth_repo.go
│   │   │   ├── branches_repo.go
│   │   │   ├── cards_repository.go
│   │   │   ├── config_versions_repo.go
│   │   │   ├── device_store_repo.go
│   │   │   ├── devices_repo.go
│   │   │   ├── iot_timeseries_repo.go
│   │   │   ├── json_util.go
│   │   │   ├── memory_tenants.go
│   │   │   ├── memory_units.go
│   │   │   ├── postgres_alarm_cloud_integration_test.go
│   │   │   ├── postgres_alarm_cloud.go
│   │   │   ├── postgres_alarm_device_integration_test.go
│   │   │   ├── postgres_alarm_device.go
│   │   │   ├── postgres_alarm_events.go
│   │   │   ├── postgres_auth.go
│   │   │   ├── postgres_branches.go
│   │   │   ├── postgres_card.go
│   │   │   ├── postgres_cards_integration_test.go
│   │   │   ├── postgres_config_versions_integration_test.go
│   │   │   ├── postgres_config_versions.go
│   │   │   ├── postgres_device_store.go
│   │   │   ├── postgres_device_store.go.bak
│   │   │   ├── postgres_devices_integration_test.go
│   │   │   ├── postgres_devices_integration_test.go.bak
│   │   │   ├── postgres_devices_test.go.bak
│   │   │   ├── postgres_devices.go
│   │   │   ├── postgres_devices.go.bak
│   │   │   ├── postgres_iot_timeseries_integration_test.go
│   │   │   ├── postgres_iot_timeseries.go
│   │   │   ├── postgres_residents_integration_test.go
│   │   │   ├── postgres_residents.go
│   │   │   ├── postgres_residents.go.bak
│   │   │   ├── postgres_role_permissions_integration_test.go
│   │   │   ├── postgres_role_permissions.go
│   │   │   ├── postgres_roles_integration_test.go
│   │   │   ├── postgres_roles.go
│   │   │   ├── postgres_round_details_integration_test.go
│   │   │   ├── postgres_round_details.go
│   │   │   ├── postgres_rounds_integration_test.go
│   │   │   ├── postgres_rounds.go
│   │   │   ├── postgres_sleepace_reports.go
│   │   │   ├── postgres_snomed_mapping_integration_test.go
│   │   │   ├── postgres_snomed_mapping.go
│   │   │   ├── postgres_tenants_integration_test.go
│   │   │   ├── postgres_tenants.go
│   │   │   ├── postgres_units_groupList_test.go.bak
│   │   │   ├── postgres_units_groupList_test.go.skip
│   │   │   ├── postgres_units_integration_test.go
│   │   │   ├── postgres_units_test.go.bak
│   │   │   ├── postgres_units_test.go.skip
│   │   │   ├── postgres_units.go
│   │   │   ├── postgres_user_branches.go
│   │   │   ├── postgres_users_integration_test.go
│   │   │   ├── postgres_users.go
│   │   │   ├── repository.go
│   │   │   ├── residents_repo.go
│   │   │   ├── residents_repo.go.bak
│   │   │   ├── role_permissions_repo.go
│   │   │   ├── roles_repo.go
│   │   │   ├── round_details_repo.go
│   │   │   ├── rounds_repo.go
│   │   │   ├── sleepace_reports_repo.go
│   │   │   ├── snomed_mapping_repo.go
│   │   │   ├── tenant_resolver.go
│   │   │   ├── tenants_repo.go
│   │   │   ├── tenants_types.go
│   │   │   ├── units_repo.go
│   │   │   ├── user_branches_repo.go
│   │   │   └── users_repo.go
│   │   ├── service
│   │   │   ├── alarm_cloud_service_integration_test.go
│   │   │   ├── alarm_cloud_service.go
│   │   │   ├── alarm_event_service_integration_test.go
│   │   │   ├── alarm_event_service.go
│   │   │   ├── auth_service_integration_test.go
│   │   │   ├── auth_service.go
│   │   │   ├── branch_service.go
│   │   │   ├── card_manage_client.go
│   │   │   ├── card_service_test.go
│   │   │   ├── card_service_vital_focus_time.go
│   │   │   ├── card_service_vital_focus.go
│   │   │   ├── card_service.go
│   │   │   ├── constants.go
│   │   │   ├── device_monitor_settings_service_integration_test.go
│   │   │   ├── device_monitor_settings_service.go
│   │   │   ├── device_service_integration_test.go
│   │   │   ├── device_service.go
│   │   │   ├── iot_timeseries_client.go
│   │   │   ├── radar_service.go
│   │   │   ├── resident_account_settings_test.go
│   │   │   ├── resident_password_reset_integration_test.go
│   │   │   ├── resident_service_test.go
│   │   │   ├── resident_service.go
│   │   │   ├── resident_update_integration_test.go
│   │   │   ├── role_permission_service_integration_test.go
│   │   │   ├── role_permission_service.go
│   │   │   ├── role_service_integration_test.go
│   │   │   ├── role_service.go
│   │   │   ├── server.go
│   │   │   ├── sleepace_client.go
│   │   │   ├── sleepace_report_service_test.go
│   │   │   ├── sleepace_report_service.go
│   │   │   ├── unit_service_hierarchy_design.md
│   │   │   ├── unit_service_hierarchy_usage.md
│   │   │   ├── unit_service_integration_test.go
│   │   │   ├── unit_service.go
│   │   │   ├── user_service_integration_test.go
│   │   │   ├── user_service.go
│   │   │   ├── vital_focus_service.go
│   │   │   └── vital_focus_util.go
│   │   └── store
│   │       └── kv.go
│   ├── main
│   ├── pkg
│   ├── repository.test
│   ├── RESIDENT_SERVICE_DESIGN_PROPOSAL.md
│   ├── RESIDENT_SERVICE_ISSUES.md
│   ├── RESIDENT_SERVICE_REPOSITORY_INTERFACES.md
│   ├── scripts
│   │   ├── add_systemadmin_users_permission.go
│   │   ├── check_all_roles_permissions.go
│   │   ├── check_c1_assignment.go
│   │   ├── check_db_connection.sh
│   │   ├── check_permissions.go
│   │   ├── check_system_roles_permissions.go
│   │   ├── check_systemadmin_manager_permissions.go
│   │   ├── check_systemadmin_permissions.go
│   │   ├── create_sleepace_report_table.go
│   │   ├── debug_permission_logic.go
│   │   ├── delete_family_role.go
│   │   ├── diagnose_cards_data.go
│   │   ├── fix_device_store.go
│   │   ├── fix_device_store.sql
│   │   ├── fix_permissions.go
│   │   ├── init_system_users.go
│   │   ├── insert_missing_role_permissions.go
│   │   ├── insert_systemadmin_branches_permissions.go
│   │   ├── insert_systemadmin_permissions.go
│   │   ├── insert_systemoperator_branches_permissions.go
│   │   ├── migrate_permission_scope.go
│   │   ├── monitor_auth_logs.sh
│   │   ├── prepare_device_test_data.sql
│   │   ├── run_local.sh
│   │   ├── simulate_c1_full_flow.go
│   │   ├── start_server.sh
│   │   ├── test_api_format.go
│   │   ├── test_api_response.go
│   │   ├── test_assigned_only_permission.go
│   │   ├── test_auth_endpoints.sh
│   │   ├── test_c1_card_query.go
│   │   ├── test_card_overview_api.go
│   │   ├── test_card_overview_api.sh
│   │   ├── test_device_endpoints.sh
│   │   ├── test_e2e.sh
│   │   ├── update_manager_branch_only.go
│   │   └── verify_permission_scope.go
│   ├── start_with_doctor.sh
│   ├── SYSTEMADMIN_DEVICE_ACCESS_FIX.md
│   └── wisefido-data
├── wisefido-iot-timeseries
│   ├── cmd
│   │   └── wisefido-iot-timeseries
│   │       └── main.go
│   ├── docs
│   │   ├── FIELD_ORDER_ANALYSIS.md
│   │   └── STREAM_DB_FORMAT_DIFF.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── config
│   │   │   └── config.go
│   │   ├── consumer
│   │   │   └── stream_consumer.go
│   │   ├── http
│   │   │   └── handler.go
│   │   ├── publisher
│   │   │   └── stream_publisher.go
│   │   └── repository
│   │       ├── field_extractor.go
│   │       └── iot_timeseries_repo.go
│   ├── start-iot-timeseries.sh
│   └── wisefido-iot-timeseries
├── wisefido-radar
│   ├── cmd
│   │   ├── decode-track
│   │   │   ├── DB_QUERY_ISSUE.md
│   │   │   ├── decoder.go
│   │   │   ├── main.go
│   │   │   ├── reader.go
│   │   │   ├── README_DB_QUERY.md
│   │   │   └── README.md
│   │   └── wisefido-radar
│   │       └── main.go
│   ├── decode-track
│   ├── docs
│   │   └── ALARM_PROCESSING_VERIFICATION.md
│   ├── exports
│   │   ├── command_service.go
│   │   ├── config.go
│   │   ├── mqtt_consumer.go
│   │   ├── radar.go
│   │   └── subscription_manager.go
│   ├── generate-cert.sh
│   ├── go.mod
│   ├── go.sum
│   ├── HTTPS_AUTH_IMPLEMENTATION_STATUS.md
│   ├── HTTPS_AUTH_IMPROVEMENT.md
│   ├── HTTPS_AUTH_STATUS.md
│   ├── HTTPS_SERVICE_DETECTION.md
│   ├── internal
│   │   ├── alarm
│   │   │   └── device_alarm_handler.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── consumer
│   │   │   ├── config_consumer.go
│   │   │   ├── mqtt_consumer_test.go
│   │   │   └── mqtt_consumer.go
│   │   ├── http
│   │   │   ├── auth_handler.go
│   │   │   ├── auth_service.go
│   │   │   ├── command_handler.go
│   │   │   ├── command_service.go
│   │   │   ├── router.go
│   │   │   └── server.go
│   │   ├── models
│   │   │   ├── auth_request.go
│   │   │   ├── auth_response.go
│   │   │   └── mqtt_command.go
│   │   ├── mqtt
│   │   ├── ota
│   │   ├── publisher
│   │   │   └── mqtt_publisher.go
│   │   ├── repository
│   │   │   └── device.go
│   │   └── service
│   │       ├── radar.go
│   │       └── subscription_manager.go
│   ├── owlBack_tree.md
│   ├── pkg
│   │   └── mqtt
│   │       └── topic_builder.go
│   ├── PORT_CONFIG_CHECK.md
│   ├── QUICK_TEST_GUIDE.md
│   ├── RADAR_SUBSCRIPTION_ANALYSIS.md
│   ├── RADAR_SUBSCRIPTION_V1_V2_COMPARISON.md
│   ├── server.crt
│   ├── server.key
│   ├── start-radar.sh
│   ├── SUBSCRIPTION_IMPLEMENTATION.md
│   ├── SUBSCRIPTION_MIGRATION_SUMMARY.md
│   ├── TEST_RESULTS.md
│   ├── test-auth-multi.sh
│   ├── test-auth.sh
│   ├── test-https-remote.sh
│   ├── test-startup.sh
│   ├── test-subscription-detailed.sh
│   ├── test-subscription.sh
│   ├── WISEFIDO_RADAR_FUNCTIONALITY_ANALYSIS.md
│   ├── WISEFIDO_RADAR_PHASE2_SUMMARY.md
│   ├── WISEFIDO_RADAR_REQUIREMENTS.md
│   └── wisefido-radar
└── wisefido-sleepace
    ├── cmd
    │   └── wisefido-sleepace
    │       └── main.go
    ├── go.mod
    ├── go.sum
    ├── internal
    │   ├── config
    │   │   └── config.go
    │   ├── consumer
    │   │   └── mqtt_consumer.go
    │   ├── models
    │   │   └── message.go
    │   ├── mqtt
    │   ├── repository
    │   │   └── device.go
    │   └── service
    │       └── sleepace.go
    ├── pkg
    ├── start-sleepace.sh
    ├── WISEFIDO_SLEEPACE_PHASE3_SUMMARY.md
    └── wisefido-sleepace

136 directories, 645 files
