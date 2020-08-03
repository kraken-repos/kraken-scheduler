package com.sap.s4hana.eureka.business.krakenscheduler.api.claimrequest.v1.dto;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class ClaimRequestDetailLineDTO {
//    private Long id;

    private String description;

    private Long materialId;

    private BigDecimal lineMoneyAmount;

    private String lineMoneyCurrency;

    private BigDecimal unitPriceAmount;

    private String unitPriceCurrency;
}
