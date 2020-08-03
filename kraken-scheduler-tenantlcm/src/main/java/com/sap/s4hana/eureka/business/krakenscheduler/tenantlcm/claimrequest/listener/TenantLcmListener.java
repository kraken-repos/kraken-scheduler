package com.sap.s4hana.eureka.business.krakenscheduler.tenantlcm.claimrequest.listener;

import com.sap.s4hana.eureka.framework.common.ApplicationContextHolder;
import com.sap.s4hana.eureka.framework.event.annotation.Consumer;
import com.sap.s4hana.eureka.framework.event.annotation.Property;
import com.sap.s4hana.eureka.framework.event.constant.EventConstant;
import com.sap.s4hana.eureka.framework.event.message.Message;
import com.sap.s4hana.eureka.framework.event.receiver.DefaultMessageListnerWithPlugin;
import com.sap.s4hana.eureka.framework.event.receiver.context.JobExecContext;
import com.sap.s4hana.eureka.framework.tenant.lcm.TenantLcmTaskExecutor;
import lombok.extern.slf4j.Slf4j;

@Slf4j
public class TenantLcmListener extends DefaultMessageListnerWithPlugin {
    private static final String KAFKA_TOPIC_TENANT_LCM = "TenantEvent";

    private TenantLcmTaskExecutor tenantLcmTaskExecutor;

    @Override
    @Consumer(name = "TenantLcmConsumer",
            properties = {
                    @Property(name = EventConstant.CONFIG_TOPIC_NAME, value = KAFKA_TOPIC_TENANT_LCM)
            })
    public void doWork(Message message, JobExecContext jobExecContext) {
        getTenantLcmTaskExecutor().execute(new String(message.getBody()), null);
    }

    TenantLcmTaskExecutor getTenantLcmTaskExecutor() {
        if (tenantLcmTaskExecutor == null) {
            synchronized (this) {
                if (tenantLcmTaskExecutor == null) {
                    tenantLcmTaskExecutor = ApplicationContextHolder.getBean(TenantLcmTaskExecutor.class);
                }
            }
        }

        return tenantLcmTaskExecutor;
    }
}

