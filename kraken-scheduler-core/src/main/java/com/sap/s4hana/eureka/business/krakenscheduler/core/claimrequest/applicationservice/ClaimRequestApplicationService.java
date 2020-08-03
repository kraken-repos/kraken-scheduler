package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.applicationservice;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.repository.ClaimRequestRepository;
import com.sap.s4hana.eureka.framework.common.converter.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@Transactional
public class ClaimRequestApplicationService {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClaimRequestApplicationService.class);

    @Autowired
    ClaimRequestRepository repository;

    @Autowired
    ObjectMapper objectMapper;

    /**
     * Create claim request. It's created by uploading document, publish events for document extraction.
     * @param claimRequest
     * @return
     */
    public Long createClaimRequestByUpload(ClaimRequest claimRequest) {
        Long claimRequestId = repository.create(claimRequest);
        claimRequest.createdByUpload();

        return claimRequestId;
    }

    /**
     * auto fill the claim request when extraction document result is available from message replied. Send notification to
     * frontend if auto fill succeeds.
     * @param claimRequest the new claim request converted from the extraction result
     *
     */
    public void autoFill(ClaimRequest claimRequest) {
        ClaimRequest theClaimRequest = repository.load(claimRequest.getId());
        theClaimRequest.autoFill(claimRequest);
    }


    /**
     * Finalize a claim request. Update claim request first and create a new claim from the claim request. The status of claim request
     * is supposed to be changed to finalized
     * @param id         the id of claim request to be finalized
     * @param claimRequest  the claim request to be finalized
     * @return the id of generated claim
     */
    public void finalize(Long id, ClaimRequest claimRequest) {
        // update the deduction
        ClaimRequest theClaimRequest = repository.load(id);
        theClaimRequest.doFinalize(claimRequest);
    }

    public ClaimRequest get(Long id) {
        return repository.load(id);
    }

    /**
     * List claim requests that has not been finalized yet, sorted by id desc.
     * @param pageSize
     * @param pageNo
     * @return
     */
    public List<ClaimRequest> listNotFinalized(Integer pageSize, Integer pageNo) {
        return repository.findNotFinalized(pageSize, pageNo);
    }

    /**
     * Get claim request by claim rquest number
     * @param claimRequestNumber
     * @return
     */
    public ClaimRequest findByClaimRequestNumber(String claimRequestNumber) {
        return repository.loadByBusinessKey(claimRequestNumber);
    }

    /**
     * Get claim requests created within last 3 days
     * @param pageSize
     * @param pageNo
     * @return
     */
    public List<ClaimRequest> listLast3Days(Integer pageSize, Integer pageNo) {
        return repository.findLast3Days(pageSize, pageNo);
    }
}
