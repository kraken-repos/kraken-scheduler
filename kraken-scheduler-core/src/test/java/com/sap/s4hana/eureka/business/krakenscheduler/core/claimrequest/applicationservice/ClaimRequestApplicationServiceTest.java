package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.applicationservice;

import static org.mockito.Mockito.*;
import static org.junit.jupiter.api.Assertions.*;

import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.repository.ClaimRequestRepository;
import com.sap.s4hana.eureka.framework.common.converter.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;


@ExtendWith(MockitoExtension.class)
public class ClaimRequestApplicationServiceTest {

    @Mock
    private ClaimRequestRepository repositoryMock;

    @Mock
    private ObjectMapper objectMapperMock;

    @InjectMocks
    private ClaimRequestApplicationService target;

    @BeforeEach
    void setUp() {
    }

    @Test
    void createClaimRequestByUpload_will_initialize_the_request() {
        // arrange
        ClaimRequest claimRequest = mock(ClaimRequest.class);

        // act
        target.createClaimRequestByUpload(claimRequest);

        // assert
        verify(claimRequest).createdByUpload();
        verify(repositoryMock).create(claimRequest);
    }

    @Test
    void createClaimRequestByUpload_will_call_repository_to_create_the_claim_request() {
        // arrange
        final long expectedId = 29L;

        ClaimRequest claimRequest = mock(ClaimRequest.class);

        when(repositoryMock.create(any(ClaimRequest.class)))
                .thenReturn(expectedId);

        // act
        Long requestId = target.createClaimRequestByUpload(claimRequest);

        // assert
        assertEquals(expectedId, requestId);

        verify(repositoryMock).create(same(claimRequest));
    }

    @Test
    void autoFill_will_load_claim_request_from_repository_by_id() {
        // arrange
        final long id = 19L;
        ClaimRequest claimToAutoFill = buildClaimRequestWithId(id);
        setupRepositoryLoadById(id);

        // act
        target.autoFill(claimToAutoFill);

        // assert
        verify(repositoryMock).load(eq(id));
    }

    @Test
    void autoFill_will_map_claim_request_to_the_request_loaded_from_repository() {
        // arrange
        final long id = 19L;
        ClaimRequest claimToAutoFill = buildClaimRequestWithId(id);
        ClaimRequest claimFromRepo = setupRepositoryLoadById(id);

        // act
        target.autoFill(claimToAutoFill);

        // assert
        verify(claimFromRepo).autoFill(same(claimToAutoFill));
    }

    @Test
    void finalize_will_load_claim_request_from_repository_by_id() {
        // arrange
        final long id = 9L;
        ClaimRequest claimToFinalize = mock(ClaimRequest.class);
        ClaimRequest loadedClaim = setupRepositoryLoadById(id);

        // act
        target.finalize(id, claimToFinalize);

        // assert
        verify(repositoryMock).load(eq(id));
        verify(loadedClaim).doFinalize(same(claimToFinalize));
    }

    @Test
    void get_will_delegate_to_dependency() {
        // arrange
        final long id = 7L;

        ClaimRequest expectedClaimRequest = setupRepositoryLoadById(id);

        // act
        ClaimRequest claimRequest = target.get(id);

        // assert
        assertSame(expectedClaimRequest, claimRequest);

        verify(repositoryMock).load(eq(id));
    }

    @Test
    void findByClaimRequestNumber_will_delegate_to_dependency() {
        // arrange
        final String claimRequestNumber = "CR_001";

        var claimRequestMock = mock(ClaimRequest.class);

        when(repositoryMock.loadByBusinessKey(claimRequestNumber))
                .thenReturn(claimRequestMock);

        // act
        ClaimRequest claimRequest = target.findByClaimRequestNumber(claimRequestNumber);

        // assert
        assertNotNull(claimRequest);
        assertSame(claimRequestMock, claimRequest);
    }

    @Test
    void listLast3Days_will_delegate_to_dependency() {
        // arrange
        final Integer pageSize = 5;
        final Integer pageNo = 3;

        var claimRequestListMock = (List<ClaimRequest>) mock(List.class);

        when(repositoryMock.findLast3Days(pageSize, pageNo))
                .thenReturn(claimRequestListMock);

        // act
        List<ClaimRequest> claimRequestList = target.listLast3Days(pageSize, pageNo);

        // assert
        assertNotNull(claimRequestList);
        assertSame(claimRequestListMock, claimRequestList);
    }

    private ClaimRequest setupRepositoryLoadById(long id) {
        ClaimRequest claimRequest = mock(ClaimRequest.class);

        when(repositoryMock.load(any(Long.class)))
                .thenReturn(claimRequest);

        return claimRequest;
    }

    private ClaimRequest buildClaimRequestWithId(long id) {
        ClaimRequest claimRequest = mock(ClaimRequest.class);

        when(claimRequest.getId()).thenReturn(id);

        return claimRequest;
    }
}
