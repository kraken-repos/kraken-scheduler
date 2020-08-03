package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.listener;

import static org.mockito.Mockito.*;

import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.converter.DebitNotesExcelExtractionConverter;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.DocumentExtractResult;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.applicationservice.ClaimRequestApplicationService;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.framework.common.ApplicationContextHolder;
import com.sap.s4hana.eureka.framework.common.utils.JsonUtils;
import com.sap.s4hana.eureka.framework.event.constant.MQMessageConst;
import com.sap.s4hana.eureka.framework.event.message.Message;
import com.sap.s4hana.eureka.framework.event.receiver.context.JobExecContext;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.BeanFactory;

public class DocumentExtractListenerTest {

    private DebitNotesExcelExtractionConverter converterMock;
    private ClaimRequestApplicationService serviceMock;

    private DocumentExtractListener target;

    @BeforeEach
    void setUp() {
        BeanFactory beanFactoryMock = mock(BeanFactory.class);

        ApplicationContextHolder appContext = new ApplicationContextHolder();

        appContext.setBeanFactory(beanFactoryMock);

        converterMock = mock(DebitNotesExcelExtractionConverter.class);

        when(beanFactoryMock.getBean(DebitNotesExcelExtractionConverter.class))
                .thenReturn(converterMock);

        serviceMock = mock(ClaimRequestApplicationService.class);

        when(beanFactoryMock.getBean(ClaimRequestApplicationService.class))
                .thenReturn(serviceMock);

        target = new DocumentExtractListener();
    }

    @Test
    void doWork_will_call_all_these_depencencies() {
        // arrange
        DocumentExtractResult docResult = new DocumentExtractResult();

        docResult.setStatus("INITIAL");
        docResult.setClaimRequestId(12345L);
        docResult.setDocumentType("DOC");
        docResult.setCountry("US");

        String msgId = "MSG_001";

        Message msg = prepareMessage(msgId, docResult);

        ClaimRequest requestMock = mock(ClaimRequest.class);

        when(converterMock.convert2Deduction(any(String.class), any(DocumentExtractResult.class)))
                .thenReturn(requestMock);

        // act
        target.doWork(msg, new JobExecContext());

        // assert
        verify(converterMock).convert2Deduction(eq(msgId), eq(docResult));
        verify(serviceMock).autoFill(same(requestMock));
    }

    private Message prepareMessage(String msgId, DocumentExtractResult bodyObj) {
        Message msgMock = mock(Message.class);

        when(msgMock.getHeader(MQMessageConst.H_MESSAGEID, String.class))
                .thenReturn(msgId);
        when(msgMock.getBody())
                .thenReturn(JsonUtils.toBytes(bodyObj));

        return msgMock;
    }
}
