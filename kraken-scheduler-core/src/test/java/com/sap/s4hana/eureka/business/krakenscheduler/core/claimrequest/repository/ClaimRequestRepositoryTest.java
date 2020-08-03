package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.repository;

import static org.mockito.Mockito.*;
import static org.junit.jupiter.api.Assertions.*;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.query.Criteria;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.query.Operator;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.query.Order;
import com.sap.s4hana.eureka.framework.rds.object.common.bo.query.Path;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import java.util.List;

public class ClaimRequestRepositoryTest {

    private ClaimRequestRepository target;

    @BeforeEach
    void setUp() {
        target = Mockito.spy(new ClaimRequestRepository());
    }

    @Test
    void findOrderBy_will_call_find_and_return_found_list() {
        // arrange
        final Integer pageSize = 10;
        final Integer pageNo = 1;
        List<ClaimRequest> expectedList = mock(List.class);

        doReturn(expectedList)
                .when(target).find(any(Criteria.class));

        // act
        List<ClaimRequest> list = target.findNotFinalized(pageSize, pageNo);

        // assert
        assertSame(expectedList, list);
    }

    @Test
    void findOrderBy_will_call_find_with_expected_criteria() {
        // arrange
        final Integer pageSize = 10;
        final Integer pageNo = 1;
        List<ClaimRequest> retList = mock(List.class);
        CriteriaHolder criteriaHolder = new CriteriaHolder();

        doAnswer(iom -> {
            criteriaHolder.setCriteria((Criteria) iom.getArgument(0));

            return retList;
        }).when(target).find(any(Criteria.class));

        // act
        target.findNotFinalized(pageSize, pageNo);

        // assert
        assertNotNull(criteriaHolder.getCriteria());

        Criteria criteria = criteriaHolder.getCriteria();
        assertEquals(1, criteria.getOrderList().size());

        Order order = criteria.getOrderList().get(0);
        assertNotNull(order.getExpression());
        assertTrue(order.getExpression() instanceof Path);
    }

    @Test
    void findLast3Days_will_call_find_with_expected_criteria() {
        // arrange
        final Integer pageSize = 5;
        final Integer pageNo = 3;

        var retList = (List<ClaimRequest>) mock(List.class);
        var criteriaHolder = new CriteriaHolder();

        doAnswer(iom -> {
            criteriaHolder.setCriteria((Criteria) iom.getArgument(0));

            return retList;
        }).when(target).find(any(Criteria.class));

        // act
        List<ClaimRequest> claimRequestList = target.findLast3Days(pageSize, pageNo);

        // assert
        assertSame(retList, claimRequestList);
        assertNotNull(criteriaHolder.getCriteria());

        Criteria criteria = criteriaHolder.getCriteria();

        assertEquals(pageSize.longValue(), criteria.getTop());
        assertEquals(((pageNo-1) * pageSize), criteria.getSkip());
        assertTrue(criteria.getRestriction() instanceof Operator.GreaterThanOrEqual);
    }

    private static class CriteriaHolder {
        private Criteria criteria;

        public Criteria getCriteria() {
            return criteria;
        }

        public void setCriteria(Criteria criteria) {
            this.criteria = criteria;
        }
    }
}
