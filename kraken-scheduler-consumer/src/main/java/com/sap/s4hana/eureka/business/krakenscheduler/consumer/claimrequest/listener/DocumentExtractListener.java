package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.listener;

import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.constant.ListenerConstants;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.converter.DebitNotesExcelExtractionConverter;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.applicationservice.ClaimRequestApplicationService;
import com.sap.s4hana.eureka.framework.common.ApplicationContextHolder;
import com.sap.s4hana.eureka.framework.common.utils.JsonUtils;
import com.sap.s4hana.eureka.framework.event.annotation.Consumer;
import com.sap.s4hana.eureka.framework.event.annotation.Property;
import com.sap.s4hana.eureka.framework.event.constant.EventConstant;
import com.sap.s4hana.eureka.framework.event.constant.MQMessageConst;
import com.sap.s4hana.eureka.framework.event.message.Message;
import com.sap.s4hana.eureka.framework.event.receiver.DefaultMessageListener;
import com.sap.s4hana.eureka.framework.event.receiver.context.JobExecContext;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.DocumentExtractResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class DocumentExtractListener extends DefaultMessageListener {

    private static final Logger LOGGER = LoggerFactory.getLogger(DocumentExtractListener.class);

    private DebitNotesExcelExtractionConverter debitNotesExcelExtractionConverter;

    @Override
    @Consumer(name = "DocumentExtractedConsumer",
            properties = {
                    @Property(name= EventConstant.CONFIG_TOPIC_NAME, value= ListenerConstants.KAFKA_TOPIC_DOC_EXTRACT)
            })
    public void doWork(Message message, JobExecContext jobExecContext) {
        String messageId = message.getHeader(MQMessageConst.H_MESSAGEID, String.class);
        DocumentExtractResult documentExtractResult = null;
        try {
            documentExtractResult = JsonUtils.fromJson(message.getBody(), DocumentExtractResult.class);
        } catch (Exception e) {
            LOGGER.error("get errors while convert document extract result json to document", e);
        }

        //get the beans. TODO change the xml declaration to annotation declaration so that beans can be autowired
        debitNotesExcelExtractionConverter = ApplicationContextHolder.getBean(DebitNotesExcelExtractionConverter.class);

        ClaimRequest claimRequest = debitNotesExcelExtractionConverter.convert2Deduction(messageId, documentExtractResult);
        ClaimRequestApplicationService claimRequestService = ApplicationContextHolder.getBean(ClaimRequestApplicationService.class);
        claimRequestService.autoFill(claimRequest);

    }
}
