// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package admin

import (
	"context"
	"errors"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/text"
)

func (p *Processor) createDomainSilence(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domain string,
	obfuscate bool,
	publicComment string,
	privateComment string,
	subscriptionID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	// Check if an allow already exists for this domain.
	domainSilence, err := p.state.DB.GetDomainSilence(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Something went wrong in the DB.
		err = gtserror.Newf("db error getting domain silence %s: %w", domain, err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	if domainSilence == nil {
		// No silence exists yet, create it.
		domainSilence = &gtsmodel.DomainSilence{
			ID:                 id.NewULID(),
			Domain:             domain,
			CreatedByAccountID: adminAcct.ID,
			PrivateComment:     text.StripHTMLFromText(privateComment),
			PublicComment:      text.StripHTMLFromText(publicComment),
			Obfuscate:          &obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// Insert the new silence into the database.
		if err := p.state.DB.PutDomainSilence(ctx, domainSilence); err != nil {
			err = gtserror.Newf("db error putting domain silence %s: %w", domain, err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}
	}

	// Run admin action to process
	// side effects of allow.
	action := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryDomain,
		TargetID:       domainSilence.Domain,
		Type:           gtsmodel.AdminActionSilence,
		AccountID:      adminAcct.ID,
	}

	if errWithCode := p.state.AdminActions.Run(
		ctx,
		action,
		// what are the side effects for domain silence?
		func(ctx context.Context) gtserror.MultiError {
			return nil
		},
	); errWithCode != nil {
		return nil, action.ID, errWithCode
	}

	apiDomainSilence, errWithCode := p.apiDomainPerm(ctx, domainSilence, false)
	if errWithCode != nil {
		return nil, action.ID, errWithCode
	}

	return apiDomainSilence, action.ID, nil
}

func (p *Processor) updateDomainSilence(
	ctx context.Context,
	domainSilenceID string,
	obfuscate *bool,
	publicComment *string,
	privateComment *string,
	subscriptionID *string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	domainSilence, err := p.state.DB.GetDomainSilenceByID(ctx, domainSilenceID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err = gtserror.Newf("db error getting domain silence: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// There are just no entries for this ID.
		err = fmt.Errorf("no domain silence entry exists with ID %s", domainSilenceID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	var columns []string
	if obfuscate != nil {
		domainSilence.Obfuscate = obfuscate
		columns = append(columns, "obfuscate")
	}
	if publicComment != nil {
		domainSilence.PublicComment = *publicComment
		columns = append(columns, "public_comment")
	}
	if privateComment != nil {
		domainSilence.PrivateComment = *privateComment
		columns = append(columns, "private_comment")
	}
	if subscriptionID != nil {
		domainSilence.SubscriptionID = *subscriptionID
		columns = append(columns, "subscription_id")
	}

	// Update the domain allow.
	if err := p.state.DB.UpdateDomainSilence(ctx, domainSilence, columns...); err != nil {
		err = gtserror.Newf("db error updating domain silence: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPerm(ctx, domainSilence, false)
}

func (p *Processor) deleteDomainSilence(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domainSilenceID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	domainSilence, err := p.state.DB.GetDomainSilenceByID(ctx, domainSilenceID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err = gtserror.Newf("db error getting domain silence: %w", err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}

		// There are just no entries for this ID.
		err = fmt.Errorf("no domain silence entry exists with ID %s", domainSilenceID)
		return nil, "", gtserror.NewErrorNotFound(err, err.Error())
	}

	// Prepare the domain silence to return, *before* the deletion goes through.
	apiDomainSilence, errWithCode := p.apiDomainPerm(ctx, domainSilence, false)
	if errWithCode != nil {
		return nil, "", errWithCode
	}

	// Delete the original domain allow.
	if err := p.state.DB.DeleteDomainSilence(ctx, domainSilence.Domain); err != nil {
		err = gtserror.Newf("db error deleting domain silence: %w", err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	// Run admin action to process
	// side effects of unsilence.
	action := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryDomain,
		TargetID:       domainSilence.Domain,
		Type:           gtsmodel.AdminActionUnsilence,
		AccountID:      adminAcct.ID,
	}

	if errWithCode := p.state.AdminActions.Run(
		ctx,
		action,
		// what are the side effects for domain unsilence?
		func(ctx context.Context) gtserror.MultiError {
			return nil
		},
	); errWithCode != nil {
		return nil, action.ID, errWithCode
	}

	return apiDomainSilence, action.ID, nil
}
