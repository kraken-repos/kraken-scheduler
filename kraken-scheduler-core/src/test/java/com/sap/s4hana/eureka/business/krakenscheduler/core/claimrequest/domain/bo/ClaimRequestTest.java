package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import static org.mockito.Mockito.*;
import static org.junit.jupiter.api.Assertions.*;

import com.sap.s4hana.eureka.framework.common.ApplicationContextHolder;
import com.sap.s4hana.eureka.framework.common.converter.ObjectMapper;
import com.sap.s4hana.eureka.framework.common.exception.BusinessException;
import com.sap.s4hana.eureka.framework.rds.object.event.service.DomainEventService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.BeanFactory;
import org.springframework.context.ApplicationContext;

import java.math.BigDecimal;
import java.util.Arrays;

public class ClaimRequestTest {

    private ApplicationContext appContextMock;
    private BeanFactory beanFactoryMock;

    @BeforeEach
    void setUp() {
        appContextMock = mock(ApplicationContext.class);
        beanFactoryMock = mock(BeanFactory.class);

        ApplicationContextHolder appContextHolder = new ApplicationContextHolder();

        appContextHolder.setApplicationContext(appContextMock);
        appContextHolder.setBeanFactory(beanFactoryMock);
    }

    @Test
    void createdByUpload_will_set_relevant_fields_property() {
        // arrange
        ClaimRequest target = ClaimRequest.builder()
                                        .status(ClaimRequestStatus.UPLOADED)
                                        .build();

        DomainEventService domainEventServiceMock = mock(DomainEventService.class);

        when(beanFactoryMock.getBean(DomainEventService.class)).thenReturn(domainEventServiceMock);

        // act
        target.createdByUpload();

        // assert
        assertSame(DocumentType.DEBIT_NOTE, target.getDocumentType());
        assertSame(ClaimRequestStatus.WAITING_FOR_AUTOFILL, target.getStatus());

        verify(domainEventServiceMock).publish(same(target), same(ClaimRequest.EVENT_TYPE_CREATEDBYUPLOAD));
    }

    @Test
    void createdByUpload_will_throw_exception_when_current_claim_request_is_finalized() {
        // arrange
        ClaimRequest target = ClaimRequest.builder()
                                        .status(ClaimRequestStatus.FINALIZED)
                                        .build();

        // act
        BusinessException exception = assertThrows(
                                            BusinessException.class,
                                            () -> target.createdByUpload());

        // assert
        String exMsg = "Finalized claim request cannot send event to extraction service";

        assertEquals(exMsg, exception.getMessage());
    }

    @Test
    void doFinalize_will_set_relevant_fields_property_and_send_event() {
        // arrange
        ClaimRequest target = ClaimRequest.builder()
                                        .status(ClaimRequestStatus.UPLOADED)
                                        .build();
        ClaimRequest inClaimRequest = mock(ClaimRequest.class);

        ObjectMapper objMapperMock = mock(ObjectMapper.class);
        DomainEventService domainEventServiceMock = mock(DomainEventService.class);

        when(beanFactoryMock.getBean(ObjectMapper.class)).thenReturn(objMapperMock);
        when(beanFactoryMock.getBean(DomainEventService.class)).thenReturn(domainEventServiceMock);

        // act
        target.doFinalize(inClaimRequest);

        // assert
        assertSame(ClaimRequestStatus.FINALIZED, target.getStatus());

        verify(objMapperMock).map(same(inClaimRequest), same(target));
        verify(domainEventServiceMock).publish(same(target), eq(ClaimRequest.EVENT_TYPE_FINALIZED));
    }

    @Test
    void doFinalize_will_throw_exception_when_current_claim_request_is_finalized() {
        // arrange
        ClaimRequest target = ClaimRequest.builder()
                                        .status(ClaimRequestStatus.FINALIZED)
                                        .build();

        // act
        BusinessException ex = assertThrows(
                                    BusinessException.class,
                                    () -> target.doFinalize(mock(ClaimRequest.class)));

        // assert
        final String expectExMsg = "This claim request is already finalized.";

        assertEquals(expectExMsg, ex.getMessage());
    }

    @Test
    void autoFill_will_set_relevant_fields_property() {
        // arrange
        ClaimRequest target = ClaimRequest.builder()
                                        .status(ClaimRequestStatus.WAITING_FOR_AUTOFILL)
                                        .build();
        ClaimRequest extractedClaimRequest = mock(ClaimRequest.class);

        ObjectMapper objMapperMock = mock(ObjectMapper.class);

        when(beanFactoryMock.getBean(ObjectMapper.class)).thenReturn(objMapperMock);

        // act
        target.autoFill(extractedClaimRequest);

        // assert
        assertSame(ClaimRequestStatus.AUTOFILLED, target.getStatus());

        verify(objMapperMock).map(extractedClaimRequest, target);
    }

    @Test
    void autoFill_will_throw_exception_when_current_claim_request_is_not_waiting_for_auto_fill() {
        // arrange
        ClaimRequest target = ClaimRequest.builder()
                                        .status(ClaimRequestStatus.UPLOADED)
                                        .build();

        // act
        BusinessException ex = assertThrows(
                                    BusinessException.class,
                                    () -> target.autoFill(mock(ClaimRequest.class)));

        // assert
        final String expectedExMsg = "The claim request can be auto filled only after it sends event to extraction service";

        assertEquals(expectedExMsg, ex.getMessage());
    }

    @Test
    void refreshTotal_will_add_all_these_amount_of_claim_lines() {
        // arrange
        final String currency = "USD";
        final MoneyV line01Money = MoneyV.of(BigDecimal.valueOf(11.11), currency);
        final MoneyV line02Money = MoneyV.of(BigDecimal.valueOf(22.11), currency);
        final MoneyV totalMoney = line01Money.add(line02Money);

        ClaimRequest target = ClaimRequest.builder()
                                        .total(MoneyV.of(BigDecimal.ZERO, currency))
                                        .detailLines(
                                                Arrays.asList(
                                                        buildClaimDetailLine(line01Money),
                                                        buildClaimDetailLine(line02Money)
                                                ))
                                        .build();

        // act
        target.refreshTotal();

        // assert
        assertEquals(totalMoney, target.getTotal());
    }

    private ClaimRequestDetailLine buildClaimDetailLine(MoneyV moneyV) {
        ClaimRequestDetailLine line = mock(ClaimRequestDetailLine.class);

        when(line.getLineMoney())
                .thenReturn(moneyV);

        return line;
    }
}
