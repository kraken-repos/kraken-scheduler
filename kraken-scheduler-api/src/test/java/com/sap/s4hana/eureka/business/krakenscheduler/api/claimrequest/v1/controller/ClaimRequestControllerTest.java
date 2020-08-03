package com.sap.s4hana.eureka.business.krakenscheduler.api.claimrequest.v1.controller;

import static org.mockito.Mockito.*;
import static org.junit.jupiter.api.Assertions.*;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.applicationservice.ClaimRequestApplicationService;
import com.sap.s4hana.eureka.business.krakenscheduler.api.claimrequest.v1.dto.ClaimRequestDTO;
import com.sap.s4hana.eureka.framework.common.converter.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

import java.util.List;

@ExtendWith(MockitoExtension.class)
public class ClaimRequestControllerTest {

    @Mock
    private ClaimRequestApplicationService serviceMock;

    @Mock
    private ObjectMapper mapperMock;

    @InjectMocks
    private ClaimRequestController target; // = new ClaimRequestController();

    @BeforeEach
    void setUp() {

    }

    @Test
    void get_will_call_dependencies() {
        // arrange
        long id = 11;
        ClaimRequest claimRequest = mock(ClaimRequest.class);
        ClaimRequestDTO claimRequestDTO = mock(ClaimRequestDTO.class);

        when(serviceMock.get(any(Long.class)))
                .thenReturn(claimRequest);
        when(mapperMock.map(any(ClaimRequest.class), any(Class.class)))
                .thenReturn(claimRequestDTO);

        // act
        ResponseEntity response = target.get(id);

        // assert
        assertNotNull(response);
        assertSame(claimRequestDTO, response.getBody());

        verify(serviceMock).get(eq((Long) id));
        verify(mapperMock).map(same(claimRequest), same(ClaimRequestDTO.class));
    }

    @Test
    void list_will_call_dependencies() {
        // arrange
        List<ClaimRequest> requestList = mock(List.class);
        List<ClaimRequestDTO> roList = mock(List.class);

        final Integer pageSize = 10;
        final Integer pageNo = 1;

        when(serviceMock.listNotFinalized(pageSize, pageNo))
                .thenReturn(requestList);
        when(mapperMock.map(any(List.class), any(Class.class)))
                .thenReturn(roList);

        // act
        ResponseEntity response = target.list(pageSize, pageNo);

        // assert
        assertNotNull(response);
        assertSame(roList, response.getBody());

        verify(serviceMock).listNotFinalized(eq(pageSize), eq(pageNo));
        verify(mapperMock).map(same(requestList), same(ClaimRequestDTO.class));
    }

    @Test
    void create_will_call_dependencies() {
        // arrange
        Long newId = 22L;
        ClaimRequestDTO requestRO = mock(ClaimRequestDTO.class);
        ClaimRequest request = mock(ClaimRequest.class);

        when(mapperMock.map(any(ClaimRequestDTO.class), any(Class.class)))
                .thenReturn(request);
        when(serviceMock.createClaimRequestByUpload(any(ClaimRequest.class)))
                .thenReturn(newId);

        // act
        ResponseEntity response = target.create(requestRO);

        // assert
        assertNotNull(response);
        assertEquals(newId, response.getBody());

        verify(mapperMock).map(same(requestRO), same(ClaimRequest.class));
        verify(serviceMock).createClaimRequestByUpload(same(request));
    }

    @Test
    void finalize_will_call_dependencies() {
        // arrange
        Long finalizeId = 33L;
        ClaimRequestDTO requestRO = new ClaimRequestDTO();
        ClaimRequest request = mock(ClaimRequest.class);

        when(mapperMock.map(any(ClaimRequestDTO.class), any(Class.class)))
                .thenReturn(request);

        // act
        target.finalize(finalizeId, requestRO);

        // assert
        verify(mapperMock).map(same(requestRO), same(ClaimRequest.class));
        verify(serviceMock).finalize(eq(finalizeId), same(request));
    }

    @Test
    void getByNumber_will_call_dependencies() {
        // arrange
        final String number = "R_00001";

        var claimRequestMock = mock(ClaimRequest.class);

        when(serviceMock.findByClaimRequestNumber(number))
                .thenReturn(claimRequestMock);

        var dtoMock = mock(ClaimRequestDTO.class);

        when(mapperMock.map(any(ClaimRequest.class), any(Class.class)))
                .thenReturn(dtoMock);

        // act
        ResponseEntity<ClaimRequestDTO> response = target.getByNumber(number);

        // assert
        assertNotNull(response);
        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertSame(dtoMock, response.getBody());

        verify(mapperMock).map(claimRequestMock, ClaimRequestDTO.class);
    }
}
