package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BOProperty;
import lombok.Getter;

import javax.persistence.Embeddable;

@Embeddable
@Getter
public class CompanyV {

    @BOProperty
    private Long customerId;

    @BOProperty
    private String name;

    @BOProperty
    private String streetAddress;

    @BOProperty
    private String city;

    @BOProperty
    private String state;

    @BOProperty
    private String country;

    @BOProperty
    private String zipCode;

    @BOProperty
    private String regionCode;

    public static CompanyV initializeWithExtractionResult(String name, String regionCode) {
        CompanyV company = new CompanyV();
        company.name = name;
        company.regionCode = regionCode;
        return company;
    }
}
