package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import com.sap.s4hana.eureka.framework.rds.object.common.bo.BOEnumType;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BOEnum;
import com.sap.s4hana.eureka.framework.rds.object.impl.converter.BaseEnumCodeConverter;

import javax.persistence.Converter;

@BOEnum
public enum ClaimRequestStatus implements BOEnumType {
    UPLOADED(10),
    WAITING_FOR_AUTOFILL(20),
    AUTOFILLED(30),
    FINALIZED(40),
    AUTOFILL_FAILED(35);

    private int enumCode;

    ClaimRequestStatus(int enumCode) {
        this.enumCode = enumCode;
    }

    @Override
    public int getEnumCode() {
        return enumCode;
    }

    @Converter(autoApply = true)
    public static class ConverterImpl implements BaseEnumCodeConverter<ClaimRequestStatus, Integer> {
    }
}

