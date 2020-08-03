package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.constant.Constants;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.eo.ClaimRequestEO;
import com.sap.s4hana.eureka.framework.common.ApplicationContextHolder;
import com.sap.s4hana.eureka.framework.common.converter.ObjectMapper;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BusinessKey;
import com.sap.s4hana.eureka.framework.rds.object.event.service.DomainEventService;
import com.sap.s4hana.eureka.framework.common.exception.BusinessException;
import com.sap.s4hana.eureka.framework.rds.dialect.ColumnTable;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BOImplementation;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.annotation.BOProperty;
import com.sap.s4hana.eureka.framework.rds.object.event.annotation.DomainEvent;
import com.sap.s4hana.eureka.framework.rds.object.impl.bo.DefaultBusinessObject;
import lombok.*;

import javax.persistence.*;
import java.util.Date;
import java.util.List;

@Entity
@Table
@ColumnTable
@BOImplementation(deletable = false)
@DomainEvent(topic = Constants.EVENT_TOPIC_CREATED_BY_UPLOAD, eventType = ClaimRequest.EVENT_TYPE_CREATEDBYUPLOAD, dtoClass = ClaimRequestEO.class)
@DomainEvent(topic = Constants.EVENT_TOPIC_FIANLIZATION, eventType = ClaimRequest.EVENT_TYPE_FINALIZED, dtoClass = ClaimRequestEO.class)
@Getter
@Builder
@NoArgsConstructor
@AllArgsConstructor(access = AccessLevel.PRIVATE)
public class ClaimRequest extends DefaultBusinessObject {

    public final static String EVENT_TYPE_CREATEDBYUPLOAD = "CreatedByUpload";
    public final static String EVENT_TYPE_FINALIZED = "Finalized";

    @Id
    @GeneratedValue(strategy = GenerationType.SEQUENCE)
    @BOProperty
    private Long id;

    @BusinessKey
    @BOProperty
    private String claimRequestNumber;

    @BOProperty
    private String documentNumber;

    @BOProperty
    private DocumentType documentType;

    @BOProperty
    private Date dueDate;

    @BOProperty
    @Embedded
    private CompanyV claimer;

    @BOProperty
    private Long attachmentId;

    @BOProperty
    @Embedded
    @AttributeOverrides({
            @AttributeOverride(name = "amount", column = @Column(name = "TOTALAMOUNT")),
            @AttributeOverride(name = "currency", column = @Column(name = "TOTALCURRENCY"))
    })
    private MoneyV total;

    @BOProperty
    @Embedded
    @AttributeOverrides({
            @AttributeOverride(name = "amount", column = @Column(name = "CLAIMEDAMOUNT")),
            @AttributeOverride(name = "currency", column = @Column(name = "CLAIMEDCURRENCY"))
    })
    private MoneyV claimed;

    @BOProperty
    private Date documentDate;

    @BOProperty
    private ClaimRequestStatus status;

    @BOProperty
    private String externalReference;

//    @OneToMany(mappedBy = "claimRequest", cascade = CascadeType.ALL, fetch = FetchType.LAZY, orphanRemoval = true)
    @ElementCollection
    @CollectionTable(
            name = "ClaimRequestDetailLine",
            joinColumns = @JoinColumn(name="claimRequestId")
    )
    List<ClaimRequestDetailLine> detailLines;

    public void setDetailLines(List<ClaimRequestDetailLine> detailLines) {
        this.detailLines = detailLines;
    }

    public List<ClaimRequestDetailLine> getDetailLines() {
        return detailLines;
    }

    /**
     * Intialize work when created by upload, and publish domain event.
     */
    public void createdByUpload() {
        this.documentType = DocumentType.DEBIT_NOTE;
        if(ClaimRequestStatus.FINALIZED.equals(this.status))
            throw new BusinessException("Finalized claim request cannot send event to extraction service");
        this.status = ClaimRequestStatus.WAITING_FOR_AUTOFILL;

        //publish domain event
        DomainEventService domainEventService = ApplicationContextHolder.getBean(DomainEventService.class);
        domainEventService.publish(this, ClaimRequest.EVENT_TYPE_CREATEDBYUPLOAD);
    }

    /**
     * Update the status of claim request after auto filled successfully.
     */
    public void autoFill(ClaimRequest extractedClaimRequest) {
        if(!this.status.equals(ClaimRequestStatus.WAITING_FOR_AUTOFILL))
            throw new BusinessException("The claim request can be auto filled only after it sends event to extraction service");

        ObjectMapper objectMapper = ApplicationContextHolder.getBean(ObjectMapper.class);
        objectMapper.map(extractedClaimRequest, this);
        this.status = ClaimRequestStatus.AUTOFILLED;
    }

    /**
     * Update claim request status after finalize the claim request.
     */
    public void doFinalize(ClaimRequest inClaimRequest) {
        if(this.status.equals(ClaimRequestStatus.FINALIZED))
            throw new BusinessException("This claim request is already finalized.");

        ObjectMapper objectMapper = ApplicationContextHolder.getBean(ObjectMapper.class);
        objectMapper.map(inClaimRequest, this);
        this.status = ClaimRequestStatus.FINALIZED;

        //publish domain event
        DomainEventService domainEventService = ApplicationContextHolder.getBean(DomainEventService.class);
        domainEventService.publish(this, ClaimRequest.EVENT_TYPE_FINALIZED);
    }

    /**
     * refresh total based on detail line amount
     */
    public void refreshTotal() {
        MoneyV total = this.detailLines.stream().map(l -> l.getLineMoney()).reduce(this.getTotal(), MoneyV::add);
        this.total = total;
        this.claimed = total;
    }
}
