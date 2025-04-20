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

package gtsmodel

import "time"

// DomainSilence represents a federation silence against a particular domain
type DomainSilence struct {
	ID                 string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Domain             string    `bun:",nullzero,notnull"`                                           // domain to silence. Eg. 'whatever.com'
	CreatedByAccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this silence
	CreatedByAccount   *Account  `bun:"-"`                                                           // Account corresponding to createdByAccountID
	PrivateComment     string    `bun:""`                                                            // Private comment on this silence, viewable to admins
	PublicComment      string    `bun:""`                                                            // Public comment on this silence, viewable (optionally) by everyone
	Obfuscate          *bool     `bun:",nullzero,notnull,default:false"`                             // whether the domain name should appear obfuscated when displaying it publicly
	SubscriptionID     string    `bun:"type:CHAR(26),nullzero"`                                      // if this silence was created through a subscription, what's the subscription ID?
}

func (d *DomainSilence) GetID() string {
	return d.ID
}

func (d *DomainSilence) GetCreatedAt() time.Time {
	return d.CreatedAt
}

func (d *DomainSilence) GetUpdatedAt() time.Time {
	return d.UpdatedAt
}

func (d *DomainSilence) SetUpdatedAt(i time.Time) {
	d.UpdatedAt = i
}

func (d *DomainSilence) GetDomain() string {
	return d.Domain
}

func (d *DomainSilence) GetCreatedByAccountID() string {
	return d.CreatedByAccountID
}

func (d *DomainSilence) SetCreatedByAccountID(i string) {
	d.CreatedByAccountID = i
}

func (d *DomainSilence) GetCreatedByAccount() *Account {
	return d.CreatedByAccount
}

func (d *DomainSilence) SetCreatedByAccount(i *Account) {
	d.CreatedByAccount = i
}

func (d *DomainSilence) GetPrivateComment() string {
	return d.PrivateComment
}

func (d *DomainSilence) SetPrivateComment(i string) {
	d.PrivateComment = i
}

func (d *DomainSilence) GetPublicComment() string {
	return d.PublicComment
}

func (d *DomainSilence) SetPublicComment(i string) {
	d.PublicComment = i
}

func (d *DomainSilence) GetObfuscate() *bool {
	return d.Obfuscate
}

func (d *DomainSilence) SetObfuscate(i *bool) {
	d.Obfuscate = i
}

func (d *DomainSilence) GetSubscriptionID() string {
	return d.SubscriptionID
}

func (d *DomainSilence) SetSubscriptionID(i string) {
	d.SubscriptionID = i
}

func (d *DomainSilence) GetType() DomainPermissionType {
	return DomainPermissionSilence
}

func (d *DomainSilence) IsOrphan() bool {
	return d.SubscriptionID == ""
}
