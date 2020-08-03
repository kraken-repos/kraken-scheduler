package com.sap.s4hana.eureka.business.krakenscheduler.api.claimrequest.v1.controller;

import com.sap.s4hana.eureka.business.krakenscheduler.api.claimrequest.v1.dto.ClaimRequestDTO;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.applicationservice.ClaimRequestApplicationService;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.framework.common.converter.ObjectMapper;
import com.sap.s4hana.eureka.framework.common.utils.JsonUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import javax.validation.constraints.Min;
import java.util.List;

@RequestMapping(value = {"/business/v1/claim-requests"})
@Validated
@RestController
public class ClaimRequestController {
    private static final String DEFAULT_PAGE_NO = "0";
    private static final String DEFAULT_PAGE_SIZE = "10";

    private static final Logger LOGGER = LoggerFactory.getLogger(ClaimRequestController.class);

    @Autowired
    private ClaimRequestApplicationService claimRequestService;

    @Autowired
    private ObjectMapper objectMapper;

    @RequestMapping(value = "/{id}", method = RequestMethod.GET)
    public ResponseEntity<ClaimRequestDTO> get(
            @PathVariable @Min(0) long id) {
        ClaimRequest claimRequest = claimRequestService.get(id);
        ClaimRequestDTO claimRequestDTO = objectMapper.map(claimRequest, ClaimRequestDTO.class);
        return new ResponseEntity<ClaimRequestDTO>(claimRequestDTO, HttpStatus.OK);
    }

    @RequestMapping( method = RequestMethod.GET)
    public ResponseEntity<List<ClaimRequestDTO>> list(
            @RequestParam(required = false, defaultValue = DEFAULT_PAGE_SIZE) @Min(0) Integer pageSize,
            @RequestParam(required = false, defaultValue = DEFAULT_PAGE_NO) @Min(0) Integer pageNo) {
        List<ClaimRequest> claimRequests = claimRequestService.listNotFinalized(pageSize, pageNo);

        return new ResponseEntity<>(objectMapper.map(claimRequests, ClaimRequestDTO.class), HttpStatus.OK);
    }

    @RequestMapping(value = "/get-by-number", method = RequestMethod.GET)
    public ResponseEntity<ClaimRequestDTO> getByNumber(@RequestParam String number) {
        ClaimRequest claimRequest = claimRequestService.findByClaimRequestNumber(number);
        ClaimRequestDTO claimRequestDTO = objectMapper.map(claimRequest, ClaimRequestDTO.class);
        return new ResponseEntity<ClaimRequestDTO>(claimRequestDTO, HttpStatus.OK);
    }

    @RequestMapping(value = "/get-last-three-days", method = RequestMethod.GET)
    public ResponseEntity<List<ClaimRequestDTO>> getLast3Days(
            @RequestParam(required = false, defaultValue = DEFAULT_PAGE_SIZE) @Min(0) Integer pageSize, 
            @RequestParam(required = false, defaultValue = DEFAULT_PAGE_NO) @Min(0) Integer pageNo) {
        List<ClaimRequest> claimRequests = claimRequestService.listLast3Days(pageSize, pageNo);
        return new ResponseEntity<>(objectMapper.map(claimRequests, ClaimRequestDTO.class), HttpStatus.OK);
    }

    @RequestMapping(method = RequestMethod.POST)
    public ResponseEntity<Long> create(
            @RequestBody ClaimRequestDTO claimRequestDTO) {
        ClaimRequest claimRequest = objectMapper.map(claimRequestDTO, ClaimRequest.class);
        Long id = claimRequestService.createClaimRequestByUpload(claimRequest);
        return new ResponseEntity<>(id, HttpStatus.CREATED);
    }

    @RequestMapping(value = "/{id}/finalize", method = RequestMethod.POST)
    public ResponseEntity<Void> finalize(
            @PathVariable @Min(0) Long id,
            @RequestBody ClaimRequestDTO claimRequestDTO) {
        claimRequestDTO.setId(id);
        String deductionString = JsonUtils.toJson(claimRequestDTO);
        LOGGER.info("Deduction fianlize. Payload: {}", deductionString);
        ClaimRequest claimRequest = objectMapper.map(claimRequestDTO, ClaimRequest.class);
        claimRequestService.finalize(id, claimRequest);
        return new ResponseEntity<>(HttpStatus.OK);
    }
    
}

