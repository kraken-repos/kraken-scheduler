package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.eo;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequestStatus;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.DocumentType;
import lombok.Data;

import java.math.BigDecimal;
import java.util.Date;
import java.util.List;

@Data
public class ClaimRequestEO {

    private Long id;

    private String documentNumber;

    private DocumentType documentType;

    private ClaimRequestStatus status;

    private String externalReference;

    private Date dueDate;

    private Long attachmentId;

    private BigDecimal totalAmount;

    private String totalCurrency;

    private BigDecimal claimedAmount;

    private String claimedCurrency;

    private Date documentDate;

    private Long claimerCustomerId;

    private String claimerName;

    private String claimerStreetAddress;

    private String claimerCity;

    private String claimerState;

    private String claimerCountry;

    private String claimerZipCode;

    private String claimerRegionCode;

    private List<ClaimRequestDetailLineEO> detailLines;
}
