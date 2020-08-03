package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.converter;

import static org.mockito.Mockito.*;
import static org.junit.jupiter.api.Assertions.*;

import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.constant.ListenerConstants;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.DocumentExtractResult;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.Extraction;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.LineItem;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequestDetailLine;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequestStatus;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.CompanyV;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.MoneyV;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import java.math.BigDecimal;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public class DebitNotesExcelExtractionConverterTest {

    private static String NET_SETTLEMENT_AMOUNT = "net_settlement_amount";
    private static String ARTICLE_DESCRIPTION   = "article_description";
    private static String CASE_UPC              = "case_upc";
    private static String CUSTOMER              = "customer";
    private static String DC_REGION             = "dc_region";
    private static String SETTLEMENT_CURRENCY   = "settlement_currency";
    private static String SETTLEMENT_DATE       = "settlement_date";
    private static String DEBIT_NOTE            = "debit_note";
    private static String EXTERNAL_NUMBER       = "external_number";

    private DocumentExtractResult docExtractResult;
    private Extraction extraction;

    private DebitNotesExcelExtractionConverter target;

    @BeforeEach
    void setUp() {
        docExtractResult = mock(DocumentExtractResult.class);
        extraction = mock(Extraction.class);

        when(docExtractResult.getExtraction()).thenReturn(extraction);

        target = Mockito.spy(new DebitNotesExcelExtractionConverter());
    }

    @Test
    void convertExtractionLineFields2HashMap_will_filter_out_all_useless_line() {
        // arrange
        List<LineItem> items = new ArrayList<>();

        items.add(buildLineItem("uom.2", "uom.2 value"));
        items.add(buildLineItem(CASE_UPC, "case_upc value"));

        // act
        Map<String, String> map = target.convertExtractionLineFields2HashMap(items);

        // assert
        assertFalse(map.containsKey("uom.2"));
        assertTrue(map.containsKey(CASE_UPC));
    }

    @Test
    void convert2DeductionDetails_will_map_each_document_item_to_claim_request_line() {
        // arrange
        final String caseUpc1 = "L_01";
        final String caseUpc2 = "L_02";
        final String claimCurrency = "USD";
        final BigDecimal netAmount1 = new BigDecimal("101.01");
        final BigDecimal netAmount2 = new BigDecimal("202.02");
        final String description1 = "description value 1";
        final String description2 = "description value 2";
        final MoneyV claimTotal = MoneyV.of(BigDecimal.ZERO, claimCurrency);

        ClaimRequest claimRequest = mock(ClaimRequest.class);

        when(claimRequest.getTotal()).thenReturn(claimTotal);

        List<List<LineItem>> lineItems = new ArrayList<>();

        lineItems.add(buildLineItems(caseUpc1, netAmount1.toString(), description1));
        lineItems.add(buildLineItems(caseUpc2, netAmount2.toString(), description2));

        when(extraction.getLineItems()).thenReturn(lineItems);

        // act
        List<ClaimRequestDetailLine> claimLineList = target.convert2DeductionDetails(docExtractResult, claimRequest);

        // assert
        assertEquals(2, claimLineList.size());

        Optional<ClaimRequestDetailLine> claimLine1Opt =
                claimLineList.stream().filter(l -> caseUpc1.equals(l.getMaterialId())).findFirst();

        assertTrue(claimLine1Opt.isPresent());
        assertEquals(MoneyV.of(netAmount1, claimCurrency), claimLine1Opt.get().getLineMoney());
        assertEquals(description1, claimLine1Opt.get().getDescription());

        Optional<ClaimRequestDetailLine> claimLine2Opt =
                claimLineList.stream().filter(l -> caseUpc2.equals(l.getMaterialId())).findFirst();

        assertTrue(claimLine2Opt.isPresent());
        assertEquals(MoneyV.of(netAmount2, claimCurrency), claimLine2Opt.get().getLineMoney());
        assertEquals(description2, claimLine2Opt.get().getDescription());
    }

    @Test
    void convert2DeductionHeader_will_return_auto_fill_failed_claim_for_extract_failed_document() {
        // arrange
        when(docExtractResult.getStatus()).thenReturn(ListenerConstants.EXTRACTION_STATUS_FAILED);

        // act
        ClaimRequest claim = target.convert2DeductionHeader("msg id", docExtractResult);

        // assert
        assertNotNull(claim);
        Assertions.assertEquals(ClaimRequestStatus.AUTOFILL_FAILED, claim.getStatus());
    }

    @Test
    void convert2DeductionHeader_will_map_these_value_in_1st_line_to_claim_request_head() {
        // arrange
        final String customer = "CUS_001";
        final String dcRegion = "region";
        final String settlementCurrency = "USD";
        final String settlementDateStr = "02/19/20";
        final Date settlementDate = buildDate("2020-02-19");
        final String debitNote = "some note";
        final String externalNumber = "REF_001";

        final String anotherCustomer = "CUS_002";

        List<List<LineItem>> lineItems = new ArrayList<>();

        lineItems.add(buildLineItems(customer, dcRegion, settlementCurrency, settlementDateStr, debitNote, externalNumber));
        lineItems.add(buildLineItems(anotherCustomer, null, null, null, null, null));

        when(extraction.getLineItems()).thenReturn(lineItems);

        // act
        ClaimRequest claimRequest = target.convert2DeductionHeader("msg_id", docExtractResult);

        // assert
        assertNotNull(claimRequest);

        CompanyV company = claimRequest.getClaimer();

        assertNotNull(company);
        assertEquals(customer, company.getName());
        assertEquals(dcRegion, company.getRegionCode());
        assertEquals(settlementCurrency, claimRequest.getTotal().getCurrency());
        assertEquals(settlementDate, claimRequest.getDocumentDate());
        assertEquals(debitNote, claimRequest.getDocumentNumber());
        assertEquals(externalNumber, claimRequest.getExternalReference());
    }

    @Test
    void convert2Deduction_will_map_claim_request_and_calculate_total_amount() {
        // arrange
        final String claimCurrency = "USD";
        final MoneyV line1Money = MoneyV.of(BigDecimal.valueOf(11.1), claimCurrency);
        final MoneyV line2Money = MoneyV.of(BigDecimal.valueOf(22.2), claimCurrency);
        final MoneyV claimTotal = line1Money.add(line2Money);

        doReturn(ClaimRequest.builder()
                             .total(MoneyV.of(BigDecimal.ZERO, claimCurrency))
                             .build())
                .when(target).convert2DeductionHeader(any(String.class), any(DocumentExtractResult.class));

        List<ClaimRequestDetailLine> claimLineList = new ArrayList<>();

        claimLineList.add(buildClaimLine(line1Money));
        claimLineList.add(buildClaimLine(line2Money));

        doReturn(claimLineList)
                .when(target).convert2DeductionDetails(any(DocumentExtractResult.class), any(ClaimRequest.class));

        // act
        ClaimRequest claimRequest = target.convert2Deduction("msg_id", docExtractResult);

        // assert
        assertNotNull(claimRequest);
        assertEquals(claimTotal, claimRequest.getTotal());
    }

    private ClaimRequestDetailLine buildClaimLine(MoneyV moneyV) {
        ClaimRequestDetailLine claimLine = mock(ClaimRequestDetailLine.class);

        when(claimLine.getLineMoney()).thenReturn(moneyV);

        return claimLine;
    }

    private Date buildDate(String dateString) {
        SimpleDateFormat dateFormat = new SimpleDateFormat("yyyy-MM-dd");
        Date date = null;

        try {
            date = dateFormat.parse(dateString);
        } catch (Exception ex) {
        }

        return date;
    }

    private List<LineItem> buildLineItems(
            String customer,
            String dcRegion,
            String settlementCurrency,
            String settlementDate,
            String debitNote,
            String externalNumber) {
        List<LineItem> itemList = new ArrayList<>();

        itemList.add(buildLineItem(CUSTOMER, customer));
        itemList.add(buildLineItem(DC_REGION, dcRegion));
        itemList.add(buildLineItem(SETTLEMENT_CURRENCY, settlementCurrency));
        itemList.add(buildLineItem(SETTLEMENT_DATE, settlementDate));
        itemList.add(buildLineItem(DEBIT_NOTE, debitNote));
        itemList.add(buildLineItem(EXTERNAL_NUMBER, externalNumber));

        return itemList;
    }

    private List<LineItem> buildLineItems(
            String caseUpc,
            String netSettlementAmount,
            String articleDescription) {
        List<LineItem> itemList = new ArrayList<>();

        itemList.add(buildLineItem(CASE_UPC, caseUpc));
        itemList.add(buildLineItem(NET_SETTLEMENT_AMOUNT, netSettlementAmount));
        itemList.add(buildLineItem(ARTICLE_DESCRIPTION, articleDescription));

        return itemList;
    }

    private LineItem buildLineItem(String name, String value) {
        LineItem lineItem = new LineItem();

        lineItem.setName(name);
        lineItem.setValue(value);

        return lineItem;
    }
}
