package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.Test;

public class CompanyVTest {

    @Test
    void initializeWithExtractionResult_will_create_company_value_instance_correctly() {
        // arrange
        String companyName = "Demo Corp";
        String regionCode = "US";

        // act
        CompanyV companyV = CompanyV.initializeWithExtractionResult(companyName, regionCode);

        // assert
        assertEquals(companyName, companyV.getName());
        assertEquals(regionCode, companyV.getRegionCode());
    }
}
