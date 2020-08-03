package com.sap.s4hana.eureka.business.krakenscheduler.tenantlcm.claimrequest.register;

import com.sap.s4hana.eureka.framework.common.utils.JsonUtils;
import com.sap.s4hana.eureka.framework.event.constant.MQMessageConst;
import com.sap.s4hana.eureka.framework.event.message.CloudEventMessage;
import com.sap.s4hana.eureka.framework.event.sender.service.EventPublisher;
import com.sap.s4hana.eureka.framework.tenant.lcm.register.RegisterPayload;
import com.sap.s4hana.eureka.framework.tenant.lcm.register.RegisterPublisher;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
@Slf4j
public class ClaimRequestRegisterPublisher implements RegisterPublisher {
    private final static String REGISTER_TOPIC = "TenantLcmServiceRegister";

    @Autowired
    private EventPublisher eventPublisher;

    @SneakyThrows
    @Override
    public void publish(RegisterPayload registerPayload) {
        CloudEventMessage message = new CloudEventMessage();

        message.setBody(JsonUtils.toBytes(registerPayload));

        try {
            eventPublisher.publish(message, REGISTER_TOPIC);
        } catch (Exception ex) {
            log.error("Exception: {}", ex);
        }
    }
}
