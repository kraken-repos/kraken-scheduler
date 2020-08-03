package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.eo;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class ClaimRequestDetailLineEO {
//    private Long id;

    private String description;

    private String materialId;

    private BigDecimal lineMoneyAmount;

    private String lineMoneyCurrency;

    private BigDecimal unitPriceAmount;

    private String unitPriceCurrency;
}
