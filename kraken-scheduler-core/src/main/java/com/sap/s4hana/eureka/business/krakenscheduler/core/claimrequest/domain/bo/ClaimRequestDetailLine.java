package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import com.sap.s4hana.eureka.framework.rds.dialect.ColumnTable;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BONodeImplementation;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BOProperty;
import lombok.*;

import javax.persistence.*;

//@Entity
@Embeddable
@ColumnTable
@BONodeImplementation
@Getter
@Builder
@NoArgsConstructor
@AllArgsConstructor(access = AccessLevel.PRIVATE)
public class ClaimRequestDetailLine {

    @BOProperty
    private String description;

    @BOProperty
    private String materialId;

    @BOProperty
    @AttributeOverrides({
            @AttributeOverride(name = "amount", column = @Column(name = "LINEAMOUNT")),
            @AttributeOverride(name = "currency", column = @Column(name = "LINECURRENCY"))
    })
    @Embedded
    private MoneyV lineMoney;

    @BOProperty
    @AttributeOverrides({
            @AttributeOverride(name = "amount", column = @Column(name = "UNITPRICEAMOUNT")),
            @AttributeOverride(name = "currency", column = @Column(name = "UNITPRICECURRENCY"))
    })

    @Embedded
    private MoneyV unitPrice;

}
