package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.repository;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequestStatus;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.query.*;
import com.sap.s4hana.eureka.framework.rds.object.impl.repository.DefaultBusinessObjectRepository;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

@Repository
public class ClaimRequestRepository extends DefaultBusinessObjectRepository<ClaimRequest> {

    public List<ClaimRequest> findNotFinalized(Integer pageSize, Integer pageNo) {
        Criteria criteria = new Criteria();
        criteria.orderBy(new Order(new Path(Long.class, "id"), false));
        var notFinalized = new Operator.NotEqual(new Path(ClaimRequestStatus.class, "status"), new Constant(ClaimRequestStatus.FINALIZED));
        criteria.where(notFinalized);
        criteria.top(pageSize);
        criteria.skip((pageNo-1)*pageSize);
        return this.find(criteria);
    }

    public List<ClaimRequest> findLast3Days(Integer pageSize, Integer pageNo) {
        Criteria criteria = new Criteria();
        var within3Days = new Operator.GreaterThanOrEqual(new Path(LocalDateTime.class, "internalCreationTime"), new Constant(LocalDate.now().atStartOfDay().minusDays(2)));
        criteria.where(within3Days);
        criteria.top(pageSize);
        criteria.skip(( pageNo-1)*pageSize);
        return this.find(criteria);
    }

    public ClaimRequest findByClaimRequestNumber(String claimRequestNumber) {
        return this.loadByBusinessKey(claimRequestNumber);
    }
}
