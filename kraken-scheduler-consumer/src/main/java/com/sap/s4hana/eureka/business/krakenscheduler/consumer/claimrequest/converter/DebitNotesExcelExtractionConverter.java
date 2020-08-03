package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.converter;

import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.constant.ListenerConstants;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequestDetailLine;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequestStatus;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.CompanyV;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.MoneyV;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.DocumentExtractResult;
import com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction.LineItem;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.text.ParseException;
import java.text.SimpleDateFormat;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Component
public class DebitNotesExcelExtractionConverter {
    private static final Logger LOGGER = LoggerFactory.getLogger(DebitNotesExcelExtractionConverter.class);

    /**
     * Converting OCR result to claim request, including detail lines.
     * @param messageId
     * @param result
     * @return
     */
    public ClaimRequest convert2Deduction(String messageId, DocumentExtractResult result) {
        ClaimRequest claimRequest = this.convert2DeductionHeader(messageId, result);
        claimRequest.setDetailLines(this.convert2DeductionDetails(result, claimRequest));
        claimRequest.refreshTotal();
        return claimRequest;
    }

    protected ClaimRequest convert2DeductionHeader(String messageId, DocumentExtractResult result) {
        ClaimRequest.ClaimRequestBuilder builder = ClaimRequest.builder();

        if(ListenerConstants.EXTRACTION_STATUS_FAILED.equals(result.getStatus())) {
            builder.status(ClaimRequestStatus.AUTOFILL_FAILED);
            LOGGER.info("Deduction document extraction failed. Message id: {}. Fail reason: ", messageId);
        } else {
            SimpleDateFormat dateFormat = new SimpleDateFormat("MM/dd/yy");
            builder.id(result.getClaimRequestId());

            for(List<LineItem> line : result.getExtraction().getLineItems()) {
                Map<String, String> lineFieldsMap = this.convertExtractionLineFields2HashMap(line);
                CompanyV claimer = CompanyV.initializeWithExtractionResult(lineFieldsMap.get("customer"), lineFieldsMap.get("dc_region"));

                builder.claimer(claimer);
                builder.total(MoneyV.of(BigDecimal.ZERO, lineFieldsMap.get("settlement_currency")));
                try {
                    String documentDateString = lineFieldsMap.get("settlement_date");
                    if(StringUtils.isNotBlank(documentDateString))
                        builder.documentDate(dateFormat.parse(documentDateString));
                } catch (ParseException e) {
                    LOGGER.error("Date parsing error during deduction autofill. The field is documentDate and value is {}",
                            lineFieldsMap.get("settlement_date"));
                }
                builder.documentNumber(lineFieldsMap.get("debit_note"));
                builder.externalReference(lineFieldsMap.get("external_number"));
                break;
            }
        }
        return builder.build();
    }

    protected List<ClaimRequestDetailLine> convert2DeductionDetails(DocumentExtractResult result, ClaimRequest claimRequest) {
        return result.getExtraction().getLineItems().stream().map(line -> {
            Map<String, String> lineFieldsMap = this.convertExtractionLineFields2HashMap(line);
            ClaimRequestDetailLine claimRequestDetailLine = ClaimRequestDetailLine.builder()
                    .lineMoney(MoneyV.of(new BigDecimal(lineFieldsMap.get("net_settlement_amount")), claimRequest.getTotal().getCurrency()))
                    .description(lineFieldsMap.get("article_description"))
                    .materialId(lineFieldsMap.get("case_upc"))
//                    .claimRequest(claimRequest)
                    .build();
            return claimRequestDetailLine;
        }).collect(Collectors.toList());
    }

    protected Map<String, String> convertExtractionLineFields2HashMap(List<LineItem> line) {
        return line.stream().filter(li -> !li.getName().equals("uom.2")).collect(Collectors.toMap(LineItem::getName, LineItem::getValue));
    }

}
