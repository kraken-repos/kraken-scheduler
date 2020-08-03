package com.sap.s4hana.eureka.business.krakenscheduler.tenantlcm.claimrequest.notifier;

import com.sap.s4hana.eureka.framework.common.utils.JsonUtils;
import com.sap.s4hana.eureka.framework.event.message.CloudEventMessage;
import com.sap.s4hana.eureka.framework.event.sender.service.EventPublisher;
import com.sap.s4hana.eureka.framework.tenant.lcm.TaskInfo;
import com.sap.s4hana.eureka.framework.tenant.lcm.notifier.TaskNotifier;
import com.sap.s4hana.eureka.framework.tenant.lcm.notifier.TaskResultPayload;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

@Component
@Transactional
@Slf4j
public class ClaimRequestTaskNotifier implements TaskNotifier {
    private final static String COMPLETE_TOPIC = "TenantCompleteTask";
    private final static String FAILURE_TOPIC = "TenantFailTask";

    @Autowired
    private EventPublisher eventPublisher;

    @Override
    @SneakyThrows
    public void onComplete(TaskInfo taskInfo) {
        CloudEventMessage event = new CloudEventMessage();
        var payload = new TaskResultPayload(taskInfo);
        event.setBody(JsonUtils.toBytes(payload));
        eventPublisher.publish(event, COMPLETE_TOPIC);
    }

    @Override
    @SneakyThrows
    public void onFailure(Exception e, TaskInfo taskInfo) {
        log.error(e.getMessage(), e);
        CloudEventMessage event = new CloudEventMessage();
        var payload = new TaskResultPayload(taskInfo);
        event.setBody(JsonUtils.toBytes(payload));
        eventPublisher.publish(event, FAILURE_TOPIC);
    }
}
