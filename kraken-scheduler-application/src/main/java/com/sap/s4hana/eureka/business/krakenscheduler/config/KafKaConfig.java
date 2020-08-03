package com.sap.s4hana.eureka.business.krakenscheduler.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.PropertySource;

@Configuration
@PropertySource(value = "file:./mock/kafka.properties", ignoreResourceNotFound = true)
@PropertySource(value = "file:/etc/secrets/kafka/kafka.properties", ignoreResourceNotFound = true)
public class KafKaConfig {
}
