package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import com.sap.s4hana.eureka.framework.rds.object.common.bo.BOEnumType;
import com.sap.s4hana.eureka.framework.rds.object.impl.converter.BaseEnumCodeConverter;

import javax.persistence.Converter;

public enum DocumentType implements BOEnumType {

    INVOICE(0, "Invoice"),
    DEBIT_NOTE(1, "DebitNote");

    private int enumCode;

    private String description;

    DocumentType(int enumCode, String description) {
        this.enumCode = enumCode;
        this.description = description;
    }

    @Override
    public int getEnumCode() {
        return enumCode;
    }

    public String getDescription() {
        return description;
    }

    @Converter(autoApply = true)
    public static class ConverterImpl implements BaseEnumCodeConverter<DocumentType, Integer> {
    }
}

