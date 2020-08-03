package com.sap.s4hana.eureka.business.krakenscheduler.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.PropertySource;

@Configuration
@PropertySource(value = "file:./mock/hana.properties", ignoreResourceNotFound = true)
@PropertySource(value = "file:/etc/secrets/hana/hana.properties", ignoreResourceNotFound = true)
public class DataSourceConfig {
}
