package com.sap.s4hana.eureka.business.krakenscheduler.api.claimrequest.v1.dto;


import com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo.ClaimRequest;
import org.modelmapper.ModelMapper;
import org.modelmapper.convention.MatchingStrategies;

import javax.persistence.Transient;

class ClaimRequestROTest {

//    @Test
    public void testModelMapper_ClaimRequestRO() {
        ModelMapper modelMapper = new ModelMapper();

        modelMapper.getConfiguration()
                .setMatchingStrategy(MatchingStrategies.STANDARD)
                .setFieldMatchingEnabled(true)
                .setFieldAccessLevel(org.modelmapper.config.Configuration.AccessLevel.PRIVATE)
                //skip transient destination property
                .setPropertyCondition(
                        context -> context.getMapping().getLastDestinationProperty().getAnnotation(Transient.class) == null)
                //skip null value properties in source
                .setSkipNullEnabled(true)
                .setDeepCopyEnabled(true)
                //keep only elements of the source collection
                .setCollectionsMergeEnabled(false);

//        modelMapper.createTypeMap(ClaimRequest.class, ClaimRequestDTO.class);
        modelMapper.createTypeMap(ClaimRequestDTO.class, ClaimRequest.class);
//        modelMapper.createTypeMap(ClaimRequestDetailLine.class, ClaimRequestDetailLineDTO.class);
//        modelMapper.createTypeMap(ClaimRequestDetailLineDTO.class, ClaimRequestDetailLine.class);
        modelMapper.validate();
    }

}
